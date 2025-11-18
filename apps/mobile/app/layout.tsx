import type { Metadata } from "next";
import "./globals.css";
import { Toaster } from "sonner";
import { AuthProvider } from "@/lib/auth";
import { ThemeProvider } from "@/lib/theme/ThemeContext";

export const metadata: Metadata = {
  title: "Home Hub Mobile",
  description: "Mobile dashboard for Home Hub - your household information at a glance",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="antialiased">
        <AuthProvider>
          <ThemeProvider>
            {children}
            <Toaster />
          </ThemeProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
