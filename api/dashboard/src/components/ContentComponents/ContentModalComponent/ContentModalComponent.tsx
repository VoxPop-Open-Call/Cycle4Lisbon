import React from "react";

import { CButton, CModal, CModalBody } from "@coreui/react";

import {
  ContentProps,
  acceptContent,
  rejectContent,
} from "../../../Controllers/ContentControllers/ContentApi";
import { stateColumnBadge } from "../../common/TableComponent/TableComponentHelperFunctions";

import styles from "./contentModalComponent.module.scss";

interface ContentModalComponentProps {
  contentDetails: ContentProps;
  open: boolean;
  closeFunction: React.Dispatch<
    React.SetStateAction<{
      contentDetails: ContentProps;
      open: boolean;
    }>
  >;
}

const ContentModalComponent = ({
  contentDetails,
  open,
  closeFunction,
}: ContentModalComponentProps): JSX.Element => {
  const onAccept = (): void => {
    acceptContent(contentDetails)
      .then(() => {
        window.alert("Content updated");
        closeFunction((currentState) => ({ ...currentState, open: false }));
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  };

  const onReject = (): void => {
    rejectContent(contentDetails)
      .then(() => {
        window.alert("Content updated");
        closeFunction((currentState) => ({ ...currentState, open: false }));
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  };

  return (
    <CModal
      alignment="center"
      size="lg"
      visible={open}
      onClose={() =>
        closeFunction((currentState) => ({ ...currentState, open: false }))
      }
    >
      <CModalBody>
        <div className={styles.defaultDivStyle}>
          <div className={styles.titleStyle}>{contentDetails.title}</div>
          {stateColumnBadge(contentDetails.state)}
        </div>
        <div className={styles.subTitleStyle}>{contentDetails.subtitle}</div>
        <div className={styles.imageDivStyle}>
          <img src={contentDetails.imageUrl} className={styles.imageStyle} />
        </div>

        <div>
          <a href={contentDetails.articleUrl} target="_blank" rel="noreferrer">
            Article Link
          </a>
        </div>
        <div className={styles.defaultDivStyle}>
          <div>Language </div>
          <div>{contentDetails.language?.name}</div>
        </div>
        <div className={styles.buttonDivStyle}>
          <CButton
            onClick={() =>
              closeFunction((currentState) => ({
                ...currentState,
                open: false,
              }))
            }
          >
            Close
          </CButton>
          <CButton onClick={() => onReject()}>Reject</CButton>
          <CButton onClick={() => onAccept()}>Approve</CButton>
        </div>
      </CModalBody>
    </CModal>
  );
};

export default ContentModalComponent;
