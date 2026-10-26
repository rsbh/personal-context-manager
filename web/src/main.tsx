import "@raystack/apsara/style.css";
import { Theme } from "@raystack/apsara";
import { QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router";
import { App } from "./App";
import { newQueryClient, setupClient } from "./lib/query";

setupClient();
const queryClient = newQueryClient();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <Theme isRoot hasBackground defaultValue={{ appearance: "system", accentColor: "indigo" }}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </QueryClientProvider>
    </Theme>
  </StrictMode>,
);
