import { writable } from 'svelte/store';
import { navigate } from './router';

export const isAuthenticated = writable<boolean>(false);

/**
 * handleLogout clears the session both on the backend and frontend.
 * It resets the isAuthenticated store and navigates to the login page.
 */
export async function handleLogout() {
  try {
    // Call backend to clear authentication session
    if (window.go?.main?.App) {
      await window.go.main.App.Logout();
    }
    // Update frontend state
    isAuthenticated.set(false);
    // Navigate to login page
    navigate('/login');
    // Safety delay: 20ms is roughly 1 frame at 60fps.
    // This is enough to let Svelte/Browser commit the route change without being perceptible to users.
    await new Promise((resolve) => setTimeout(resolve, 20));
  } catch (error) {
    console.error('Logout failed:', error);
  }
}
