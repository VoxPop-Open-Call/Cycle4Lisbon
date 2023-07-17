import React from "react";

import { CCard, CCardBody, CCardHeader, CCol, CRow } from "@coreui/react";
import { CChart } from "@coreui/react-chartjs";

import { MetricsProps } from "../../../Controllers/AnalyticsController/AnalyticsApi";

import styles from "./userMetrics.module.scss";

interface UserMetricsProps {
  metrics: { loaded: boolean; data: MetricsProps };
}

const UserMetrics: React.FC<UserMetricsProps> = ({ metrics }): JSX.Element => {
  const renderTripMetrics = (): JSX.Element => {
    if (!metrics.loaded || !Object.keys(metrics.data.users).length) {
      return <></>;
    }
    return (
      <>
        <div className={styles.pageSubTitleStyle}>User Metrics</div>
        <CRow md={{ cols: 3 }}>
          <CCol>
            <CCard className={styles.cardStyle}>
              <CCardHeader className={styles.cardHeaderStyle}>
                Registered User&apos;s Age
              </CCardHeader>
              <CCardBody>
                <CChart
                  type="bar"
                  data={{
                    labels: [
                      "<18",
                      "18 to 25",
                      "26 to 30",
                      "31 to 40",
                      "41 to 60",
                      "61 to 75",
                      ">75",
                    ],

                    datasets: [
                      {
                        backgroundColor: [
                          "#C93046",
                          "#FDA943",
                          "#FDCA43",
                          "#E0D62E",
                          "#A2B414",
                        ],
                        data: [
                          metrics.data.users.ageGroups["age<18"],
                          metrics.data.users.ageGroups["18<=age<25"],
                          metrics.data.users.ageGroups["25<=age<30"],
                          metrics.data.users.ageGroups["30<=age<40"],
                          metrics.data.users.ageGroups["40<=age<60"],
                          metrics.data.users.ageGroups["60<=age<75"],
                          metrics.data.users.ageGroups["age>=75"],
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
            <CCard className={styles.cardStyle}>
              <CCardHeader className={styles.cardHeaderStyle}>
                Registered User&apos;s Gender
              </CCardHeader>
              <CCardBody>
                <CChart
                  type="bar"
                  data={{
                    labels: ["Male", "Female", "Other"],

                    datasets: [
                      {
                        backgroundColor: [
                          "#C93046",
                          "#FDA943",
                          "#FDCA43",
                          "#E0D62E",
                          "#A2B414",
                        ],
                        data: [
                          metrics.data.users.genderCount.m,
                          metrics.data.users.genderCount.f,
                          metrics.data.users.genderCount.x,
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
        </CRow>
      </>
    );
  };

  return renderTripMetrics();
};

export default UserMetrics;
