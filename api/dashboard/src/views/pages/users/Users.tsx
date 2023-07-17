import React, { useCallback, useEffect, useState } from "react";

import { CCard, CCardBody } from "@coreui/react";

import UsersTableComponent from "../../../components/UsersComponents/UsersTableComponent/UsersTableComponent";
import {
  UserProps,
  getUserList,
} from "../../../Controllers/UserControllers/UsersApi";

import styles from "./users.module.scss";

const Users: React.FC = (): JSX.Element => {
  const [userList, setUserList] = useState<UserProps[]>([]);
  const [pagination, setPagination] = useState({
    limit: 10,
    offset: 0,
    orderBy: "id asc",
  });
  const userListFetch = useCallback(() => {
    getUserList(pagination)
      .then(({ data }) => {
        const _data = data.map((obj) => ({
          ...obj,
          name: obj.name ? obj.name : "",
          email: obj.email ? obj.email : "",
          _props: { className: styles.tableRowStyle },
        }));
        setUserList(_data);
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  }, [pagination]);

  useEffect(() => {
    userListFetch();
  }, [userListFetch]);

  return (
    <CCard>
      <CCardBody className={styles.containingCard}>
        <div className={styles.pageTitle}>Users</div>
        <div className={styles.tableDiv}>
          <UsersTableComponent
            tableData={userList}
            pagination={pagination}
            setPagination={setPagination}
          />
        </div>
      </CCardBody>
    </CCard>
  );
};

export default Users;
