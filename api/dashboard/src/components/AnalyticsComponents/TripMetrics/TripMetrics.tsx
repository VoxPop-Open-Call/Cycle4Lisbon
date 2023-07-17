import React from "react";

import { CCol, CRow } from "@coreui/react";

import { MetricsProps } from "../../../Controllers/AnalyticsController/AnalyticsApi";
import MetricCardComponent from "../../common/TableComponent/MetricCardComponent/MetricCardComponent";

import styles from "./tipMetrics.module.scss";
import { NameMap } from "./TripMetricsHelperFunction";

interface TripMetricsProps {
  metrics: { loaded: boolean; data: MetricsProps };
}
const TripMetrics: React.FC<TripMetricsProps> = ({ metrics }): JSX.Element => {
  const renderCardText = (name: string, value: number): string | number => {
    switch (name) {
      case "total":
        return value;
      case "averageDist":
        return `${Number(value.toFixed(2))} km`;
      default:
        return Number(value).toFixed(2);
    }
  };

  const renderTripMetrics = (): JSX.Element => {
    if (!metrics.loaded || !Object.keys(metrics.data.trips).length) {
      return <></>;
    }
    return (
      <>
        <div className={styles.pageSubTitleStyle}>Trip Metrics</div>
        <CRow md={{ cols: 3 }}>
          {Object.entries(metrics.data.trips).map(([name, value]) => (
            <CCol key={name} className={styles.metricColStyle}>
              <MetricCardComponent
                cardStyle={styles.metricsCardStyle}
                cardBodyStyle={styles.metricsCardBodyStyle}
                title={NameMap[name]}
                titleStyle={styles.metricsTitleStyle}
                value={renderCardText(name, value)}
                valueStyle={styles.metricsValueStyle}
              />
            </CCol>
          ))}
        </CRow>
      </>
    );
  };

  return renderTripMetrics();
};

export default TripMetrics;
