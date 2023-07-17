import React, { Suspense, useState } from "react";

import * as Sentry from "@sentry/react";
import { useRoutes } from "react-router-dom";

import "./App.css";
import "./scss/style.scss";

import "@coreui/coreui/dist/css/coreui.min.css";
import "@coreui/coreui-pro/dist/css/coreui.min.css";
import "bootstrap/dist/css/bootstrap.min.css";

import routes from "./routes";

const useSentryRoutes = Sentry.wrapUseRoutes(useRoutes);

const App = (): JSX.Element => {
  const [userIsAuthenticated, setUserIsAuthenticated] = useState(
    localStorage.getItem("encodedToken") !== null
  );

  const routing = useSentryRoutes(
    routes(userIsAuthenticated, setUserIsAuthenticated)
  );

  return <Suspense>{routing}</Suspense>;
};

export default Sentry.withProfiler(App);
