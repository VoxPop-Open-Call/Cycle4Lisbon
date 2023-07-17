import React from "react";

import { CCard, CCardBody } from "@coreui/react";
import { useNavigate } from "react-router";

import { InitiativesProps } from "../../../Controllers/InitiativesControllers/InitiativesApi";

import styles from "./initiativesDetails.module.scss";

interface InitiativesDetailsProps {
  details: InitiativesProps;
}

const InitiativesDetails: React.FC<InitiativesDetailsProps> = ({
  details,
}): JSX.Element => {
  const navigate = useNavigate();
  const renderBreadCrumb = (): JSX.Element => (
    <div className={styles.headerContainerStyle}>
      <div className={styles.breadcrumbContainerStyle}>
        <div
          className={styles.breadcrumbActionStyle}
          onClick={(e) => {
            e.preventDefault();
            navigate("/initiatives");
          }}
        >
          Initiatives
        </div>
        <div className={styles.breadcrumbCurrentStyle}>
          {" "}
          / Initiative Detail
        </div>
      </div>
      {/* {approveRejectButtons()} */}
    </div>
  );

  const {
    title,
    description,
    goal,
    credits,
    endDate,
    institution,
    presignedImageURL,
  } = details;

  return (
    <div>
      {renderBreadCrumb()}
      <CCard className={styles.cardStyle}>
        <CCardBody className={styles.cardBodyStyle}>
          <div className={styles.contentTitleDiv}>{title}</div>
          <div className={styles.contentSubtitleContainerDiv}>
            <div className={styles.contentSubTitleDiv}>
              <span className={styles.contentFieldStyle}>Goal: </span>
              {goal} credits
            </div>
            <div className={styles.contentSubTitleDiv}>
              <span className={styles.contentFieldStyle}>
                Collected so far:{" "}
              </span>
              {Number(credits).toFixed(0)} credits
            </div>
            <div className={styles.contentSubTitleDiv}>
              <span className={styles.contentFieldStyle}>End Date: </span>
              {endDate}
            </div>
          </div>
          <div className={styles.imageStyle}>
            <img src={presignedImageURL} />
          </div>

          <div className={styles.descriptionStyle}>{description}</div>
          <div className={styles.endorseContainerDiv}>
            <div className={styles.endorseText}>Endorsed by</div>
            <img
              src={institution.presignedLogoURL}
              className={styles.endorseImg}
            />
          </div>
        </CCardBody>
      </CCard>
    </div>
  );
};

export default InitiativesDetails;
