import React from "react";

import { CCard, CCardBody } from "@coreui/react";

import styles from "./metricCardComponent.module.scss";

interface MetricCardComponentProps {
  icon?: JSX.Element;
  cardStyle?: string;
  cardBodyStyle?: string;
  title: string | number;
  value: string | number;
  titleStyle?: string;
  valueStyle?: string;
  color?: string;
}

const MetricCardComponent = ({
  icon,
  cardStyle,
  cardBodyStyle,
  title,
  value,
  titleStyle,
  valueStyle,
  color,
}: MetricCardComponentProps): JSX.Element => {
  return (
    <CCard className={cardStyle} style={{ backgroundColor: color }}>
      <CCardBody className={cardBodyStyle}>
        <div className={styles.titleValueContainerDiv}>
          <div>{icon}</div>
          <div>
            <div className={titleStyle}>{title}</div>
            <div className={valueStyle}>{value}</div>
          </div>
        </div>
      </CCardBody>
    </CCard>
  );
};

export default MetricCardComponent;
