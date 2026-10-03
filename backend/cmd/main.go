package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"biameet.ir/api"
	"biameet.ir/db"
	"biameet.ir/services"
	"biameet.ir/version"
	"biameet.ir/web"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	port := env("PORT", "8080")

	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version") {
		fmt.Println(version.Version)
		return
	}

	// `biameet healthcheck` lets the Docker HEALTHCHECK work in a scratch image (no curl).
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://127.0.0.1:" + port + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	log.SetFlags(log.LstdFlags | log.LUTC)

	log.Printf("biameet %s starting", version.Version)

	dbPath := env("DB_PATH", "biameet.db")
	if err := dropPrivileges(dbPath); err != nil {
		log.Fatalf("privileges: %v", err)
	}
	if err := db.InitDB(dbPath); err != nil {
		if strings.Contains(err.Error(), "readonly") || strings.Contains(err.Error(), "unable to open") {
			log.Printf("hint: the process (uid %d) needs write access to %s and its directory", os.Getuid(), dbPath)
		}
		log.Fatalf("database: %v", err)
	}

	// Signs the vote tokens browsers keep instead of passwords. Generated once
	// and stored in the database unless SECRET_KEY is set.
	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		var err error
		if secret, err = db.Setting("token_secret", func() string { return db.RandomHex(32) }); err != nil {
			log.Fatalf("token secret: %v", err)
		}
	}
	key, err := hex.DecodeString(secret)
	if err != nil || len(key) < 16 {
		key = []byte(secret)
	}
	services.SetTokenKey(key)

	site, err := web.New(web.Options{
		Dir:     os.Getenv("WEB_DIR"),
		BaseURL: env("BASE_URL", "https://www.biameet.ir"),
		Version: version.Version,
		// From Google Search Console → Add property → HTML tag (the content="…" value).
		GoogleVerification: os.Getenv("GOOGLE_SITE_VERIFICATION"),
	})
	if err != nil {
		log.Fatalf("frontend: %v", err)
	}
	// Compressing the assets at startup allocates scratch memory that is never
	// needed again; hand it back to the OS instead of keeping it resident.
	debug.FreeOSMemory()

	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		log.Printf("ADMIN_TOKEN not set: /api/v1/admin/stats is disabled")
	}
	writeLimit, _ := strconv.Atoi(env("WRITE_RATE_LIMIT", "30"))

	app := api.NewApp(api.Config{
		Site:        site,
		AdminToken:  adminToken,
		ProxyHeader: os.Getenv("PROXY_HEADER"),
		LogRequests: os.Getenv("LOG_REQUESTS") == "1",
		WriteLimit:  writeLimit,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("biameet %s listening on :%s", version.Version, port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("shutdown: %v", err)
	}
	if err := db.Close(); err != nil {
		log.Printf("closing database: %v", err)
	}
}
