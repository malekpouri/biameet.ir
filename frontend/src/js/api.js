const API_BASE = '/api/v1';

export class ApiError extends Error {
  constructor(code, message, status) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

// api() returns the parsed JSON body, or throws an ApiError whose message is
// the server's user-facing (Persian) text and whose code is the stable error code.
export async function api(method, path, body, headers = {}) {
  let res;
  try {
    res = await fetch(API_BASE + path, {
      method,
      headers: body === undefined ? headers : { 'Content-Type': 'application/json', ...headers },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new ApiError('network', 'ارتباط با سرور برقرار نشد. اتصال اینترنت خود را بررسی کنید.', 0);
  }

  let data = null;
  try {
    data = await res.json();
  } catch {
    /* empty or non-JSON body */
  }
  if (!res.ok) {
    throw new ApiError(
      (data && data.error) || `http_${res.status}`,
      (data && data.message) || 'خطایی رخ داد، لطفاً دوباره تلاش کنید',
      res.status,
    );
  }
  return data;
}
