import React from "react";

import EmptyIcon from "../../../assets/emptyViews/empty-view.svg";

import styles from "./pageNotImplemented.module.scss";

const PageNotImplemented = (): JSX.Element => {
  return (
    <div className={styles.containerDiv}>
      <img src={EmptyIcon} />
      <div className={styles.titleStyle}>Ride faster, see more!</div>
      <div className={styles.subtitleStyle}>
        We are working to show you more features to discover and explore while
        supporting meaningful initiatives.
      </div>
    </div>
  );
};

export default PageNotImplemented;
