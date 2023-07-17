import React from "react";

import moment from "moment";
import { useNavigate } from "react-router";

import { ContentProps } from "../../../Controllers/ContentControllers/ContentApi";
import TableComponent from "../../common/TableComponent/TableComponent";
import {
  contentColumns,
  stateColumnBadge,
} from "../../common/TableComponent/TableComponentHelperFunctions";

import styles from "./contentTableComponent.module.scss";

interface ContentTableComponentProps {
  tableData: Array<ContentProps>;
  pagination: { limit: number; offset: number; orderBy: string };
  setPagination: React.Dispatch<
    React.SetStateAction<{
      limit: number;
      offset: number;
      orderBy: string;
      type?: string;
    }>
  >;
}

const ContentTableComponent: React.FC<ContentTableComponentProps> = ({
  tableData,
  pagination,
  setPagination,
}: ContentTableComponentProps) => {
  const navigate = useNavigate();

  const columnProps = {
    date: (content: ContentProps): JSX.Element => (
      <td>{moment(content.date).format("MMMM DD, YYYY")}</td>
    ),
    state: (content: { state: string }): JSX.Element =>
      stateColumnBadge(content.state),
    actions: (content: ContentProps): JSX.Element => (
      <td>
        <div
          className={styles.buttonStyle}
          onClick={() => {
            navigate(`/news/${content.id}`, {
              state: { newsContent: content },
            });
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
        tableColumnsVisibility={contentColumns}
        pagination={pagination}
        setPagination={setPagination}
      />
    </>
  );

  return <div>{renderDataTable()}</div>;
};

export default ContentTableComponent;
