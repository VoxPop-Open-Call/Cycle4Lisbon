import React, { useCallback, useEffect, useState } from "react";

import { CCard, CCardBody } from "@coreui/react";

import InitiativesTableComponent from "../../../components/InitiativesComponents/InitiativesTableComponent/InitiativesTableComponent";
import {
  InitiativesProps,
  getInitiativesList,
} from "../../../Controllers/InitiativesControllers/InitiativesApi";

import styles from "./initiatives.module.scss";

const Initiatives: React.FC = (): JSX.Element => {
  const [initiatives, setInitiatives] = useState<InitiativesProps[]>([]);
  const [pagination, setPagination] = useState({
    limit: 10,
    offset: 0,
    orderBy: "id asc",
  });

  const initiativesListFetch = useCallback(() => {
    getInitiativesList(pagination)
      .then(({ data }) => {
        const _data = data.map((obj) => ({
          ...obj,
          title: obj?.title ? obj.title : "",
          _props: { className: styles.tableRowStyle },
        }));
        setInitiatives(_data);
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  }, [pagination]);

  useEffect(() => {
    initiativesListFetch();
  }, [initiativesListFetch]);

  return (
    <CCard>
      <CCardBody className={styles.containingCard}>
        <div className={styles.pageTitle}>Initiatives</div>
        <div className={styles.tableDiv}>
          <InitiativesTableComponent
            tableData={initiatives}
            pagination={pagination}
            setPagination={setPagination}
          />
        </div>
      </CCardBody>
    </CCard>
  );
};

export default Initiatives;
