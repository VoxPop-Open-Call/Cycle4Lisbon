import React from "react";

import { CCard, CCardBody, CCardHeader } from "@coreui/react";
import { useNavigate } from "react-router";

import { ReactComponent as ApproveIcon } from "../../../assets/unverifiedButtonsIcons/approve-icon.svg";
import { ReactComponent as RejectIcon } from "../../../assets/unverifiedButtonsIcons/reject-icon.svg";
import UserInfoCardComponent from "../../../components/UsersComponents/UserInfoCardComponent/UserInfoCardComponent";
import {
  UserProps,
  deleteUser,
  verifyUser,
} from "../../../Controllers/UserControllers/UsersApi";

import styles from "./userDetails.module.scss";

interface UserDetailsProps {
  userDetails: UserProps;
}

const UserDetails: React.FC<UserDetailsProps> = ({
  userDetails,
}): JSX.Element => {
  const navigate = useNavigate();
  const { name, username, email, verified } = userDetails;

  const onApprove = (): void => {
    verifyUser(userDetails)
      .then(() => {
        window.alert("User updated.");
        navigate("/users");
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  };

  const onDelete = (): void => {
    deleteUser(userDetails)
      .then(() => {
        window.alert("User deleted");
        navigate("/users");
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  };

  const approveRejectButtons = (): JSX.Element => {
    if (verified) {
      return <></>;
    }
    return (
      <div className={styles.buttonContainerDiv}>
        <div className={styles.acceptButtonStyle} onClick={() => onApprove()}>
          <ApproveIcon className={styles.buttonIconStyle} />
          Approve
        </div>
        <div className={styles.rejectButtonStyle} onClick={() => onDelete()}>
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
            navigate("/users");
          }}
        >
          Users
        </div>
        <div className={styles.breadcrumbCurrentStyle}> / User Detail</div>
      </div>
      {approveRejectButtons()}
    </div>
  );

  return (
    <>
      {renderBreadCrumb()}
      <CCard className={styles.cardStyle}>
        <CCardHeader className={styles.cardHeaderStyle}>
          <UserInfoCardComponent userDetails={userDetails} />
        </CCardHeader>
        <CCardBody className={styles.cardBodyStyle}>
          <div className={styles.userDetailsContainerDiv}>
            <div className={styles.userDetailsFieldsDiv}>
              <div>First Name</div>
              <div>Last Name</div>
              <div>Nickname</div>
              <div>Email</div>
            </div>
            <div className={styles.userDetailsValuesDiv}>
              <div>{name ? name.split(" ")[0] : ""}</div>
              <div>{name ? name.split(" ")[1] : ""}</div>
              <div>{username}</div>
              <div>{email}</div>
            </div>
          </div>
        </CCardBody>
      </CCard>
    </>
  );
};

export default UserDetails;
