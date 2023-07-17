import React from "react";

import { Navigate, Outlet, RouteObject } from "react-router-dom";

const Analytics = React.lazy(() => import("./views/pages/analytics/Analytics"));

const ContentDetailsManager = React.lazy(
  () => import("./views/pages/contentManagement/ContentDetailsManager")
);
const ContentManagement = React.lazy(
  () => import("./views/pages/contentManagement/ContentManagement")
);
// const Dashboard = React.lazy(() => import("./views/pages/dashboard/Dashboard"));
const FaqManager = React.lazy(() => import("./views/pages/FAQ/FaqManager"));

const Initiatives = React.lazy(
  () => import("./views/pages/Initiatives/Initiatives")
);

const InitiativesDetailsManager = React.lazy(
  () => import("./views/pages/Initiatives/InitiativesDetailsManager")
);

const Login = React.lazy(() => import("./views/pages/login/Login"));
const MainLayout = React.lazy(
  () => import("./views/pages/MainLayout/MainLayout")
);
const PageNotImplemented = React.lazy(
  () => import("./views/pages/PageNotImplemented/PageNotImplemented")
);

const PolicyManager = React.lazy(
  () => import("./views/pages/policy/PolicyManager")
);
const TermsAndConditionsManager = React.lazy(
  () => import("./views/pages/terms_and_conditions/TermsAndConditionsManager")
);
const UserDetailsManager = React.lazy(
  () => import("./views/pages/users/UserDetailsManager")
);
const Users = React.lazy(() => import("./views/pages/users/Users"));

function routes(
  isLoggedIn: boolean,
  isLoggedInFunction: React.Dispatch<React.SetStateAction<boolean>>
): Array<RouteObject> {
  return [
    {
      element: isLoggedIn ? (
        <MainLayout onLoggedChange={isLoggedInFunction} />
      ) : (
        <Navigate to="/login" />
      ),
      children: [
        { path: "/dashboard", element: <Analytics /> },
        { path: "/users", element: <Users /> },
        { path: "/user/:id", element: <UserDetailsManager /> },
        { path: "/news", element: <ContentManagement /> },
        { path: "/news/:id", element: <ContentDetailsManager /> },
        { path: "/initiatives", element: <Initiatives /> },
        { path: "/initiatives/:id", element: <InitiativesDetailsManager /> },
        { path: "/", element: <Navigate to="/dashboard" replace /> },
        { path: "*", element: <PageNotImplemented /> },
      ],
    },
    {
      element: !isLoggedIn ? <Outlet /> : <Navigate to="/dashboard" />,
      children: [
        {
          path: "login",
          element: <Login onLoggedChange={isLoggedInFunction} />,
        },
        { path: "/", element: <Navigate to="/login" /> },
      ],
    },
    {
      path: "policy",
      element: <PolicyManager />,
    },
    {
      path: "/faq",
      element: <FaqManager />,
    },
    { path: "terms-conditions", element: <TermsAndConditionsManager /> },
  ];
}

export default routes;
