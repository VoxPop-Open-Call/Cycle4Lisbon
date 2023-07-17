import React from "react";

import { cisLockUnlocked } from "@coreui/icons-pro";
import CIcon from "@coreui/icons-react";
import { CBadge, CNavItem, CNavLink } from "@coreui/react";
import { useLocation } from "react-router";
import { NavLink } from "react-router-dom";

import { NavGroupItemProps, NavItemProps } from "../sidebarRoutes";

import styles from "./appSidebarNav.module.scss";

interface AppSidebarNavProps {
  navChoices: Array<NavGroupItemProps | NavItemProps>;
  onLoggedChange: (value: boolean) => void;
}

const AppSidebarNav: React.FC<AppSidebarNavProps> = ({
  navChoices,
  onLoggedChange,
}) => {
  const location = useLocation();
  const navLink = (
    name: string,
    icon: JSX.Element,
    badge?: { color: string; text: string }
  ): JSX.Element => {
    return (
      <>
        {icon && icon}
        {name && <div>{name}</div>}
        {badge && (
          <CBadge color={badge.color} className="ms-auto">
            {badge.text}
          </CBadge>
        )}
      </>
    );
  };

  const navItem = (item: NavItemProps, index: number): JSX.Element => {
    const { component, name, icon, badge, ...rest } = item;
    const Component = component;
    return (
      <Component
        {...(rest.to && !rest.items ? { component: NavLink } : {})}
        key={index}
        {...rest}
      >
        {navLink(name, icon, badge)}
      </Component>
    );
  };

  const navGroup = (item: NavGroupItemProps, index: number): JSX.Element => {
    const { component, name, icon, to, items, ...rest } = item;
    const Component = component;
    return (
      <Component
        idx={String(index)}
        key={index}
        toggler={navLink(name, icon)}
        visible={location.pathname.startsWith(to)}
        {...rest}
      >
        {items?.map((_item: NavGroupItemProps | NavItemProps, _index: number) =>
          _item.items
            ? navGroup(_item as NavGroupItemProps, _index)
            : navItem(_item as NavItemProps, index)
        )}
      </Component>
    );
  };

  const logOutUser = (): void => {
    localStorage.clear();
    onLoggedChange(false);
  };

  const renderLogOutOption = (): JSX.Element => (
    <CNavItem>
      <CNavLink onClick={() => logOutUser()} className={styles.logoutDivStyle}>
        <CIcon icon={cisLockUnlocked} className="nav-icon" />
        <div className={styles.logoutStyle}>Log Out</div>
      </CNavLink>
    </CNavItem>
  );

  return (
    <React.Fragment>
      {navChoices &&
        navChoices.map(
          (choice: NavGroupItemProps | NavItemProps, index: number) =>
            choice.items
              ? navGroup(choice as NavGroupItemProps, index)
              : navItem(choice as NavItemProps, index)
        )}
      {renderLogOutOption()}
    </React.Fragment>
  );
};

export default AppSidebarNav;
