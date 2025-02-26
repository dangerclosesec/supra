// src/components/layout/ThemeToggle.tsx
import { useEffect, useState } from "react";
import { Sun, Moon } from "lucide-react";
import { Button } from "../ui/button";

const ThemeToggle = () => {
  const [isDark, setIsDark] = useState(true);

  useEffect(() => {
    // Check for system preference
    const systemPrefersDark = window.matchMedia(
      "(prefers-color-scheme: dark)"
    ).matches;

    // Set initial theme from localStorage or system preference
    const savedTheme = localStorage.getItem("theme");
    const initialIsDark = savedTheme
      ? savedTheme === "dark"
      : systemPrefersDark;

    setIsDark(initialIsDark);

    // Apply theme class to document
    if (initialIsDark) {
      document.documentElement.classList.remove("light");
    } else {
      document.documentElement.classList.add("light");
    }
  }, []);

  const toggleTheme = () => {
    setIsDark(!isDark);

    if (isDark) {
      document.documentElement.classList.add("light");
      localStorage.setItem("theme", "light");
    } else {
      document.documentElement.classList.remove("light");
      localStorage.setItem("theme", "dark");
    }
  };

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggleTheme}
      className="rounded-full"
      aria-label="Toggle theme"
    >
      {isDark ? (
        <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all" />
      ) : (
        <Moon className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all" />
      )}
    </Button>
  );
};

export default ThemeToggle;
