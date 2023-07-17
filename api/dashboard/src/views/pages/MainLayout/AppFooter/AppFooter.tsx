import React from "react";

import { CFooter } from "@coreui/react";

const AppFooter: React.FC = (): JSX.Element => {
  return (
    <CFooter className="mt-4 navbar fixed-bottom" style={{ zIndex: 1029 }}>
      <div className="ms-auto">
        <span className="me-1">Ajuda+ Copyright</span>
      </div>
    </CFooter>
  );
};

export default React.memo(AppFooter);
