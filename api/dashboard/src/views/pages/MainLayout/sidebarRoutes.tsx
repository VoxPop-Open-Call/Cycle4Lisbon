import React, { PropsWithChildren } from "react";

import {
  cisBell,
  cisChart,
  cisCog,
  cisNewspaper,
  cisPeople,
  cisSpeedometer,
} from "@coreui/icons-pro";
import CIcon from "@coreui/icons-react";
import { CNavItem } from "@coreui/react";

import "./sidebarRoutes.module.scss";

export interface NavGroupComponentProps {
  idx?: string;
  key?: string | number | null;
  toggler: JSX.Element;
  visible: boolean;
}

export interface NavItemComponentProps {
  to: string;
  items?: object;
}

export interface NavGroupItemProps {
  component: React.ComponentType<PropsWithChildren<NavGroupComponentProps>>;
  name: string;
  icon: JSX.Element;
  to: string;
  items?: Array<NavGroupItemProps | NavItemProps>;
}

export interface NavItemProps {
  component: React.ComponentType<PropsWithChildren<NavItemComponentProps>>;
  name: string;
  badge?: { color: string; text: string };
  icon: JSX.Element;
  to: string;
  items?: object;
}

const sidebarRoutes: Array<NavGroupItemProps | NavItemProps> = [
  {
    component: CNavItem,
    name: "Dashboard",
    to: "/dashboard",
    icon: <CIcon icon={cisSpeedometer} customClassName="nav-icon" />,
  },
  {
    component: CNavItem,
    name: "Users",
    to: "/users",
    icon: <CIcon icon={cisPeople} customClassName="nav-icon" />,
  },
  // {
  //   component: CNavItem,
  //   name: "Analytics",
  //   to: "/analytics",
  //   icon: <CIcon icon={cisChart} customClassName="nav-icon" />,
  // },
  {
    component: CNavItem,
    name: "Initiatives",
    to: "/initiatives",
    icon: <CIcon icon={cisChart} customClassName="nav-icon" />,
  },
  {
    component: CNavItem,
    name: "Notifications",
    to: "/notifications",
    icon: <CIcon icon={cisBell} customClassName="nav-icon" />,
  },
  {
    component: CNavItem,
    name: "News & Events",
    to: "/news",
    icon: <CIcon icon={cisNewspaper} customClassName="nav-icon" />,
  },
  {
    component: CNavItem,
    name: "Settings",
    to: "/settings",
    icon: <CIcon icon={cisCog} customClassName="nav-icon" />,
  },
];

export default sidebarRoutes;
