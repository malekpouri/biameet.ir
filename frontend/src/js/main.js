import { showAdmin } from './admin.js';
import { showCreate } from './create.js';
import { renderMessage, showSession } from './session.js';
import { $, bindDelegation, purgeLegacyPasswords } from './util.js';

purgeLegacyPasswords();

const app = $('app');
bindDelegation(app);

// The server decides which view a URL is (see backend/api/pages.go).
switch (document.body.dataset.page) {
  case 'session':
    showSession(app, location.pathname.slice(1));
    break;
  case 'admin':
    showAdmin(app);
    break;
  case 'notfound':
    renderMessage(app, 'صفحه یافت نشد', 'جلسه یا صفحه‌ای که دنبالش هستید وجود ندارد یا حذف شده است.');
    break;
  default:
    showCreate(app);
}
