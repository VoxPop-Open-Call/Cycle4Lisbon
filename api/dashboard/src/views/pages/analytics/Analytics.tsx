import React, { useCallback, useMemo, useState } from "react";

import PlatformMetrics from "../../../components/AnalyticsComponents/PlatformMetrics/PlatformMetrics";
import TripMetrics from "../../../components/AnalyticsComponents/TripMetrics/TripMetrics";
import UserMetrics from "../../../components/AnalyticsComponents/UserMetrics/UserMetrics";
import {
  MetricsProps,
  getMetrics,
} from "../../../Controllers/AnalyticsController/AnalyticsApi";

import styles from "./analytics.module.scss";

const Analytics = (): JSX.Element => {
  const [analyticsSwitch, setAnalyticsSwitch] = useState("platform");
  const [metrics, setMetrics] = useState({
    loaded: false,
    data: {} as MetricsProps,
  });

  const metricsFetch = useCallback(() => {
    getMetrics()
      .then(({ data }) => {
        setMetrics((currentValue) => ({
          ...currentValue,
          loaded: true,
          data: data,
        }));
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  }, []);

  useMemo(() => {
    metricsFetch();
  }, [metricsFetch]);

  return (
    <>
      <div className={styles.pageTitleStyle}>
        Analytics
        <div className={styles.metricTypeSelectionDiv}>
          <div
            onClick={() => setAnalyticsSwitch("platform")}
            className={
              analyticsSwitch === "platform"
                ? styles.activePlatformMetricTypeStyle
                : styles.platformMetricTypeStyle
            }
          >
            Platform
          </div>
          <div
            onClick={() => setAnalyticsSwitch("trip")}
            className={
              analyticsSwitch === "trip"
                ? styles.activeTripMetricTypeStyle
                : styles.tripMetricTypeStyle
            }
          >
            Trips
          </div>
          <div
            onClick={() => setAnalyticsSwitch("user")}
            className={
              analyticsSwitch === "user"
                ? styles.activeUserMetricTypeStyle
                : styles.userMetricTypeStyle
            }
          >
            Users
          </div>
        </div>
      </div>
      {analyticsSwitch === "platform" ? (
        <PlatformMetrics metrics={metrics} />
      ) : null}
      {analyticsSwitch === "trip" ? <TripMetrics metrics={metrics} /> : null}
      {analyticsSwitch === "user" ? <UserMetrics metrics={metrics} /> : null}
    </>
  );
};

export default Analytics;
