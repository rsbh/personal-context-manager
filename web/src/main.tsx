import "@raystack/apsara/style.css";
import "./global.css";
import { Theme, Toast } from "@raystack/apsara";
import { QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router";
import { routes } from "./App";
import { newQueryClient, setupClient } from "./lib/query";

setupClient();
const queryClient = newQueryClient();
const router = createBrowserRouter(routes);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <Theme
      isRoot
      hasBackground
      defaultValue={{ appearance: "system", accentColor: "indigo", grayColor: "slate" }}
    >
      <Toast.Provider>
        <QueryClientProvider client={queryClient}>
          <RouterProvider router={router} />
        </QueryClientProvider>
      </Toast.Provider>
    </Theme>
  </StrictMode>,
);
