import React from "react";

import * as Sentry from "@sentry/react";
import {
  Routes,
  createRoutesFromChildren,
  matchRoutes,
  useLocation,
  useNavigationType,
} from "react-router-dom";

import config from "./config";

export const InitSentry = (): void => {
  // Disable Sentry in development.
  if (
    process.env.NODE_ENV === "development" ||
    window.location.host.includes("localhost")
  ) {
    return;
  }

  Sentry.init({
    dsn: config.SENTRY_DSN,
    environment: window.location.host,

    // Sample only a small percentage of user traces
    // sampleRate: 1.0,
    tracesSampleRate: 0.01,
    replaysSessionSampleRate: 0.01,
    // If the entire session is not sampled, use the below sample rate to sample
    // sessions when an error occurs.
    replaysOnErrorSampleRate: 0.2,

    integrations: [
      new Sentry.Replay({ maskAllText: true, blockAllMedia: true }),
      new Sentry.BrowserTracing({
        // Propagate tracing id's to our API's
        tracePropagationTargets: [
          "localhost",
          "https://api.cycleforlisbon.com",
        ],
        routingInstrumentation: Sentry.reactRouterV6Instrumentation(
          React.useEffect,
          useLocation,
          useNavigationType,
          createRoutesFromChildren,
          matchRoutes
        ),
      }),
    ],
  });
};

export const sentryCreateBrowserRouter = (): React.FC =>
  Sentry.withSentryReactRouterV6Routing(Routes);
