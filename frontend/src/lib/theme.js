const STORAGE_KEY = 'tms-theme';
const GLASS_STORAGE_KEY = 'tms-glass';

export const getInitialTheme = () => {
  if (typeof window === 'undefined') return 'light';
  const stored = window.localStorage.getItem(STORAGE_KEY);
  if (stored === 'light' || stored === 'dark') return stored;
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
};

// "Glass" (tema SBG del mockup "TMS Unificato"): switch indipendente,
// persistito a parte — si compone con light o dark (vedi .glass/.glass.dark
// in index.css), non è un terzo valore esclusivo di theme.
export const getInitialGlass = () => {
  if (typeof window === 'undefined') return false;
  return window.localStorage.getItem(GLASS_STORAGE_KEY) === '1';
};

export const applyTheme = (theme, glass) => {
  document.documentElement.classList.toggle('dark', theme === 'dark');
  document.documentElement.classList.toggle('glass', !!glass);
  window.localStorage.setItem(STORAGE_KEY, theme);
  window.localStorage.setItem(GLASS_STORAGE_KEY, glass ? '1' : '0');
};
