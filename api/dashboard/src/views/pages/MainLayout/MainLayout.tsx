import React, { Suspense } from "react";

import { CContainer, CSpinner } from "@coreui/react";
import { Outlet } from "react-router";

import AppHeader from "./AppHeader/AppHeader";
import AppSidebar from "./AppSidebar/AppSidebar";

interface MainLayoutProps {
  onLoggedChange: (value: boolean) => void;
}

const MainLayout: React.FC<MainLayoutProps> = ({ onLoggedChange }) => {
  return (
    <div>
      <AppSidebar onLoggedChange={onLoggedChange} />
      <div className="wrapper d-flex flex-column min-vh-100 bg-light">
        <AppHeader />
        <div className="body flex-grow-1 px-3 mb-4">
          <CContainer className="mb-5">
            <Suspense fallback={<CSpinner color="primary" />}>
              <Outlet />
            </Suspense>
          </CContainer>
        </div>
        {/* <AppFooter /> */}
      </div>
    </div>
  );
};

export default MainLayout;
