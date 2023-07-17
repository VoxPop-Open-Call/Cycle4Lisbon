import React from "react";

import { useLocation } from "react-router";

import ContentDetails from "./ContentDetails";

const ContentDetailsManager = (): JSX.Element => {
  const { state } = useLocation();
  const { newsContent } = state;
  return <ContentDetails contentDetails={newsContent} />;
};

export default ContentDetailsManager;
