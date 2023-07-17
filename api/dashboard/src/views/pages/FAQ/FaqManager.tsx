import React from "react";

import { useLocation } from "react-router";

import FaqEN from "./FaqEN";

const FaqManager = (): JSX.Element => {
  const location = useLocation();
  const params = new URLSearchParams(location.search);
  const selectedLang = params.get("lang");

  const renderFAQ = (): JSX.Element => {
    switch (selectedLang) {
      default:
        return <FaqEN />;
    }
  };

  return renderFAQ();
};

export default FaqManager;
