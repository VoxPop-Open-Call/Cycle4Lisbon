import React from "react";

import { useNavigate } from "react-router";

import { InitiativesProps } from "../../../Controllers/InitiativesControllers/InitiativesApi";
import TableComponent from "../../common/TableComponent/TableComponent";
import { initiativeColumns } from "../../common/TableComponent/TableComponentHelperFunctions";

import styles from "./initiativesTableComponent.module.scss";
import {
  renderInstitutions,
  renderSponsors,
  renderStatusBadge,
} from "./InitiativesTableHelperFunctions";

interface InitiativesTableComponentProps {
  tableData: Array<InitiativesProps>;
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

const InitiativesTableComponent: React.FC<InitiativesTableComponentProps> = ({
  tableData,
  pagination,
  setPagination,
}): JSX.Element => {
  const navigate = useNavigate();

  const columnProps = {
    credits: (initiative: InitiativesProps) => (
      <td>{Number(initiative.credits).toFixed(0)}</td>
    ),
    enabled: (initiative: InitiativesProps): JSX.Element => (
      <td>{renderStatusBadge(initiative)}</td>
    ),
    institution: (initiative: InitiativesProps): JSX.Element => (
      <td>{renderInstitutions(initiative)}</td>
    ),
    sponsors: (initiative: InitiativesProps): JSX.Element => (
      <td>
        <div className={styles.sponsorContainerDiv}>
          {renderSponsors(initiative)}
        </div>
      </td>
    ),
    actions: (initiative: InitiativesProps) => (
      <td>
        <div
          className={styles.buttonStyle}
          onClick={() => {
            navigate(`/initiatives/${initiative.id}`);
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
        tableColumnsVisibility={initiativeColumns}
        pagination={pagination}
        setPagination={setPagination}
      />
    </>
  );

  return <div>{renderDataTable()}</div>;
};

export default InitiativesTableComponent;
