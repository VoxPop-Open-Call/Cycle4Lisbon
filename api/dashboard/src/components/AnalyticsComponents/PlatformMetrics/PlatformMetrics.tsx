import React from "react";

import { CCard, CCardBody, CCardHeader, CCol, CRow } from "@coreui/react";
import { CChart } from "@coreui/react-chartjs";

import { MetricsProps } from "../../../Controllers/AnalyticsController/AnalyticsApi";
import MetricCardComponent from "../../common/TableComponent/MetricCardComponent/MetricCardComponent";

import styles from "./platformMetrics.module.scss";

interface PlatformMetricsProps {
  metrics: { loaded: boolean; data: MetricsProps };
}

const PlatformMetrics: React.FC<PlatformMetricsProps> = ({
  metrics,
}): JSX.Element => {
  const renderPlatformMetrics = (): JSX.Element => {
    if (!metrics.loaded || !Object.keys(metrics.data.platform).length) {
      return <></>;
    }
    return (
      <>
        <div className={styles.pageSubTitleStyle}>Platform Metrics</div>
        <CRow md={{ cols: 3 }}>
          <CCol>
            <CCard className={styles.cardStyle}>
              <CCardHeader className={styles.cardHeaderStyle}>
                Initiatives
              </CCardHeader>
              <CCardBody>
                <CChart
                  type="pie"
                  data={{
                    labels: [
                      "Total Initiatives",
                      "Completed Initiatives",
                      "Ongoing Initiatives",
                    ],

                    datasets: [
                      {
                        label: "Initiatives",
                        backgroundColor: [
                          "#C93046",
                          "#FDA943",
                          "#FDCA43",
                          "#E0D62E",
                          "#A2B414",
                        ],
                        data: [
                          metrics.data.platform.totalInitiatives,
                          metrics.data.platform.completedInitiatives,
                          metrics.data.platform.ongoingInitiatives,
                        ],
                        hoverOffset: 10,
                      },
                    ],
                  }}
                  options={{ plugins: { legend: { display: false } } }}
                  customTooltips={false}
                />
              </CCardBody>
            </CCard>
          </CCol>
          <CCol>
            <MetricCardComponent
              cardStyle={styles.metricsCardStyle}
              cardBodyStyle={styles.metricsCardBodyStyle}
              title={"Total Credits Generated"}
              titleStyle={styles.metricsTitleStyle}
              value={Number(metrics.data.platform.totalCledits).toFixed(0)}
              valueStyle={styles.metricsValueStyle}
            />
          </CCol>
        </CRow>
      </>
    );
  };

  return renderPlatformMetrics();
};

export default PlatformMetrics;
