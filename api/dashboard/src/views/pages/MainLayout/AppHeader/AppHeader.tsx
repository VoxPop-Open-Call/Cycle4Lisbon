import React from "react";

import { cilMenu } from "@coreui/icons";
import CIcon from "@coreui/icons-react";
import { CContainer, CHeader, CHeaderToggler } from "@coreui/react";
import { useDispatch, useSelector } from "react-redux";

interface SidebarVisibility {
  sidebarShow: boolean;
}

const AppHeader: React.FC = (): JSX.Element => {
  const dispatch = useDispatch();
  const sidebarShow = useSelector(
    (state: SidebarVisibility) => state.sidebarShow
  );
  return (
    <CHeader position="sticky" className="mb-4">
      <CContainer fluid>
        <CHeaderToggler
          className="ps-1"
          onClick={() => dispatch({ type: "set", sidebarShow: !sidebarShow })}
        >
          <CIcon icon={cilMenu} size="lg" />
        </CHeaderToggler>
      </CContainer>
    </CHeader>
  );
};

export default AppHeader;
