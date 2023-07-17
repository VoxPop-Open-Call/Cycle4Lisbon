import React, { useCallback, useEffect, useState } from "react";

import { CCard, CCardBody } from "@coreui/react";

import ContentTableComponent from "../../../components/ContentComponents/ContentTableComponent/ContentTableComponent";
import {
  ContentProps,
  getContentList,
} from "../../../Controllers/ContentControllers/ContentApi";

import styles from "./contentManagement.module.scss";

const ContentManagement: React.FC = (): JSX.Element => {
  const [contentList, setContentList] = useState<ContentProps[]>([]);
  const [pagination, setPagination] = useState<{
    limit: number;
    offset: number;
    orderBy: string;
    type?: string;
  }>({
    limit: 10,
    offset: 0,
    orderBy: "id asc",
    type: "news",
  });

  const contentListFetch = useCallback(() => {
    getContentList(pagination)
      .then(({ data }) => {
        const _data = data.map((obj) => ({
          ...obj,
          _props: { className: styles.tableRowStyle },
        }));
        setContentList(_data);
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  }, [pagination]);

  useEffect(() => {
    contentListFetch();
  }, [contentListFetch]);

  return (
    <CCard className={styles.containingCard}>
      <CCardBody className={styles.containingCardBody}>
        <div className={styles.pageTitle}>News & Events</div>
        <div className={styles.tableDiv}>
          <ContentTableComponent
            tableData={contentList}
            pagination={pagination}
            setPagination={setPagination}
          />
        </div>
      </CCardBody>
    </CCard>
  );
};

export default ContentManagement;
