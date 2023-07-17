import React from "react";

import { useNavigate } from "react-router";

import { UserProps } from "../../../Controllers/UserControllers/UsersApi";
import TableComponent from "../../common/TableComponent/TableComponent";
import {
  userColumns,
  verifiedColumnBadge,
} from "../../common/TableComponent/TableComponentHelperFunctions";

import styles from "./usersTableComponent.module.scss";

interface UsersTableComponentProps {
  tableData: Array<UserProps>;
  pagination: {
    limit: number;
    offset: number;
    orderBy: string;
  };
  setPagination: React.Dispatch<
    React.SetStateAction<{
      limit: number;
      offset: number;
      orderBy: string;
    }>
  >;
}

const UsersTableComponent: React.FC<UsersTableComponentProps> = ({
  tableData,
  pagination,
  setPagination,
}: UsersTableComponentProps): JSX.Element => {
  const navigate = useNavigate();
  const columnProps = {
    verified: (user: { verified: boolean }): JSX.Element =>
      verifiedColumnBadge(user.verified),
    actions: (user: UserProps) => (
      <td>
        <div
          className={styles.buttonStyle}
          onClick={() => {
            navigate(`/user/${user.id}`);
          }}
        >
          Show
        </div>
      </td>
    ),
  };

  const renderDataTable = (): JSX.Element => (
    <>
      <TableComponent
        tableData={tableData}
        tableColumnProps={columnProps}
        tableColumnsVisibility={userColumns}
        pagination={pagination}
        setPagination={setPagination}
      />
    </>
  );

  return <div>{renderDataTable()}</div>;
};

export default UsersTableComponent;
