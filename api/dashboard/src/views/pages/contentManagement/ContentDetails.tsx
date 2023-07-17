import React from "react";

import { CCard, CCardBody } from "@coreui/react";
import { useNavigate } from "react-router";

import { ReactComponent as ApproveIcon } from "../../../assets/unverifiedButtonIcons/approve-icon.svg";
import { ReactComponent as RejectIcon } from "../../../assets/unverifiedButtonIcons/reject-icon.svg";
import {
  ContentProps,
  acceptContent,
  rejectContent,
} from "../../../Controllers/ContentControllers/ContentApi";

import styles from "./contentDetails.module.scss";

interface ContentDetailsProps {
  contentDetails: ContentProps;
}

const ContentDetails: React.FC<ContentDetailsProps> = ({
  contentDetails,
}): JSX.Element => {
  const navigate = useNavigate();
  const {
    title,
    subtitle,
    subject,
    period,
    time,
    state,
    imageUrl,
    description,
    articleUrl,
  } = contentDetails;

  const onAccept = (): void => {
    acceptContent(contentDetails)
      .then(() => {
        window.alert("Content updated");
        navigate("/news");
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  };

  const onReject = (): void => {
    rejectContent(contentDetails)
      .then(() => {
        window.alert("Content updated");
        navigate("/news");
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  };

  const approveRejectButtons = (): JSX.Element => {
    if (state !== "pending") {
      return <></>;
    }
    return (
      <div className={styles.buttonContainerDiv}>
        <div className={styles.acceptButtonStyle} onClick={() => onAccept()}>
          <ApproveIcon className={styles.buttonIconStyle} />
          Approve
        </div>
        <div className={styles.rejectButtonStyle} onClick={() => onReject()}>
          <RejectIcon className={styles.buttonIconStyle} />
          Reject
        </div>
      </div>
    );
  };

  const renderBreadCrumb = (): JSX.Element => (
    <div className={styles.headerContainerStyle}>
      <div className={styles.breadcrumbContainerStyle}>
        <div
          className={styles.breadcrumbActionStyle}
          onClick={(e) => {
            e.preventDefault();
            navigate("/news");
          }}
        >
          News & Events
        </div>
        <div className={styles.breadcrumbCurrentStyle}> / Content Detail</div>
      </div>
      {approveRejectButtons()}
    </div>
  );

  return (
    <>
      {renderBreadCrumb()}
      <CCard className={styles.cardStyle}>
        <CCardBody className={styles.cardBodyStyle}>
          <div className={styles.contentTitleDiv}>{title}</div>
          <div className={styles.contentSubtitleContainerDiv}>
            <div className={styles.contentSubTitleDiv}>{subtitle}</div>
            <div className={styles.contentSubTitleDiv}>{subject}</div>
            <div className={styles.contentSubTitleDiv}>{period}</div>
            <div className={styles.contentSubTitleDiv}>{time}</div>
          </div>

          <img src={imageUrl} className={styles.imageStyle} />
          <div className={styles.descriptionStyle}>{description}</div>
          <div className={styles.sourceLinkContainerDiv}>
            Source link:{" "}
            <a
              href={articleUrl}
              target="_blank"
              rel="noreferrer"
              className={styles.articleUrlStyle}
            >
              click here
            </a>
          </div>
        </CCardBody>
      </CCard>
    </>
  );
};

export default ContentDetails;
