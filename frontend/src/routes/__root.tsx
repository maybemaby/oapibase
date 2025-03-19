import {
  Link,
  Outlet,
  createRootRoute,
  createRootRouteWithContext,
} from "@tanstack/react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import * as React from "react";
import "../index.css";
import "unfonts.css";
import { AuthProvider } from "@/components/auth-provider";

interface RouteContext {
  auth: null | {
    userId: string;
  };
}

const queryClient = new QueryClient();

const TanstackRouterDevtools =
  process.env.NODE_ENV === "production"
    ? () => null // Render nothing in production
    : React.lazy(() =>
      // Lazy load in development
      import("@tanstack/router-devtools").then((res) => ({
        default: res.TanStackRouterDevtools,
        // For Embedded Mode
        // default: res.TanStackRouterDevtoolsPanel
      }))
    );

export const Route = createRootRouteWithContext<RouteContext>()({
  component: RootComponent,

});

function RootComponent() {
  return (
    <>
      <Outlet />
      <React.Suspense>
        <TanstackRouterDevtools position="bottom-right" />
      </React.Suspense>
    </>
  );
}
