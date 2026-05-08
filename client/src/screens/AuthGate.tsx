import { useEffect, type ReactNode } from "react";

import { onAuthExpired } from "@/api";
import { useAuthStore } from "@/store/authStore";

import { LoginScreen } from "./LoginScreen";

interface Props {
  children: ReactNode;
}

export function AuthGate({ children }: Props) {
  const token = useAuthStore((s) => s.token);
  const signOut = useAuthStore((s) => s.signOut);

  useEffect(() => onAuthExpired(() => signOut()), [signOut]);

  if (!token) return <LoginScreen />;
  return <>{children}</>;
}
