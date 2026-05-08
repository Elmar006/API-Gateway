import "@/styles/globals.css";
import "@xyflow/react/dist/style.css";

import { QueryClientProvider } from "@tanstack/react-query";
import React from "react";
import ReactDOM from "react-dom/client";

// Auth store registers the token provider with the API client on import.
import "@/store/authStore";

import { App } from "@/App";
import { queryClient } from "@/lib/queryClient";
import { AuthGate } from "@/screens/AuthGate";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <AuthGate>
        <App />
      </AuthGate>
    </QueryClientProvider>
  </React.StrictMode>,
);
