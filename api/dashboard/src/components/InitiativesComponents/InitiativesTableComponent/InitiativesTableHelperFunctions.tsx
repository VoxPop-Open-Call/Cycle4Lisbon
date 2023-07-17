import React from "react";

import { InitiativesProps } from "../../../Controllers/InitiativesControllers/InitiativesApi";

import styles from "./initiativesTableComponent.module.scss";

export const renderStatusBadge = (
  initiative: InitiativesProps
): JSX.Element => {
  const { enabled } = initiative;
  if (enabled) {
    return <div className={styles.acceptedBadge}>Enabled</div>;
  }
  return <div className={styles.rejectedBadge}>Disabled</div>;
};

export const renderInstitutions = (initiative: InitiativesProps): string => {
  const { institution } = initiative;
  if (!Object.keys(institution).length) {
    return "";
  }
  return institution.name;
};

const sponsorNameMap: { [key: string]: string } = {
  "Câmara Municipal de Lisboa": "CML",
};

export const renderSponsors = (initiative: InitiativesProps): JSX.Element[] => {
  if (!initiative?.sponsors?.length) {
    return [<></>];
  }
  const { title, sponsors } = initiative;
  return sponsors.map(({ name }) => (
    <div key={title + name} className={styles.sponsorBadge}>
      {sponsorNameMap[name] ? sponsorNameMap[name] : name}
    </div>
  ));
};
