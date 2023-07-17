import React from "react";

import { useLocation } from "react-router";

import TermsAndConditionsEN from "./TermsAndConditionsEN";

const TermsAndConditionsManager = (): JSX.Element => {
  const location = useLocation();
  const params = new URLSearchParams(location.search);
  const selectedLang = params.get("lang");
  const renderTermsAndServices = (): JSX.Element => {
    switch (selectedLang) {
      default:
        return <TermsAndConditionsEN />;
    }
  };
  return renderTermsAndServices();
};

export default TermsAndConditionsManager;
