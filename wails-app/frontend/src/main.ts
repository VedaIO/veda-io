import { mount } from 'svelte';
import 'bootstrap/dist/js/bootstrap.bundle.min.js';
import './global.css';
import App from './App.svelte';

console.log('Svelte app initialized with Wails');

const app = mount(App, {
  target: document.getElementById('app')!,
});

export default app;
