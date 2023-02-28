/* @refresh reload */
import { render } from 'solid-js/web';

import App from './App.jsx';

const app = document.getElementById('app');

if (import.meta.env.DEV && !(app instanceof HTMLElement)) {
  throw new Error(
    'Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got mispelled?',
  );
}

render(() => <App />, app)
