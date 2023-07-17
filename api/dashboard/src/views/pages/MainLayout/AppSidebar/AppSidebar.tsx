import React from "react";

import {
  CSidebar,
  CSidebarBrand,
  CSidebarNav,
  CSidebarToggler,
} from "@coreui/react";
import { useDispatch, useSelector } from "react-redux";
import SimpleBar from "simplebar-react";

import "simplebar-react/dist/simplebar.min.css";

import { ReactComponent as ShortLogoIcon } from "../../../../assets/sidebarIcons/logo-icon.svg";
import { ReactComponent as LongLogoIcon } from "../../../../assets/sidebarIcons/long_logo.svg";
import AppSidebarNav from "../AppSidebarNav/AppSidebarNav";
import sidebarMenuOptions from "../sidebarRoutes";

import styles from "./appSidebar.module.scss";

interface Foldable {
  sidebarUnfoldable: boolean;
}

interface SidebarVisibility {
  sidebarShow: boolean;
}

interface AppSidebarProps {
  onLoggedChange: (value: boolean) => void;
}

const AppSidebar: React.FC<AppSidebarProps> = ({ onLoggedChange }) => {
  const dispatch = useDispatch();
  const unfoldable = useSelector((state: Foldable) => state.sidebarUnfoldable);
  const sidebarShow = useSelector(
    (state: SidebarVisibility) => state.sidebarShow
  );
  return (
    <CSidebar
      className={styles.sidebarStyle}
      position="fixed"
      unfoldable={unfoldable}
      visible={sidebarShow}
      onVisibleChange={(visible) => {
        dispatch({ type: "set", sidebarShow: visible });
      }}
    >
      <CSidebarBrand className={`d-none d-md-flex ${styles.sidebarIconStyle}`}>
        <LongLogoIcon className="sidebar-brand-full" height={35} />
        <ShortLogoIcon className="sidebar-brand-narrow" height={35} />
      </CSidebarBrand>
      <CSidebarNav>
        <SimpleBar>
          <AppSidebarNav
            navChoices={sidebarMenuOptions}
            onLoggedChange={onLoggedChange}
          />
        </SimpleBar>
      </CSidebarNav>
      <CSidebarToggler
        className="d-none d-lg-flex"
        style={{ borderRadius: "0px" }}
        onClick={() =>
          dispatch({ type: "set", sidebarUnfoldable: !unfoldable })
        }
      />
    </CSidebar>
  );
};

export default AppSidebar;
