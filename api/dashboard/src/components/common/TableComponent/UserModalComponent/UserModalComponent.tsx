import React from "react";

import { CButton, CCol, CModal, CModalBody, CRow } from "@coreui/react";
import moment from "moment";

import {
  UserProps,
  deleteUser,
  verifyUser,
} from "../../../../Controllers/UserControllers/UsersApi";

import styles from "./userModalComponent.module.scss";

interface UserModalComponentProps {
  userDetails: UserProps;
  open: boolean;
  closeFunction: React.Dispatch<
    React.SetStateAction<{
      userDetails: UserProps;
      open: boolean;
    }>
  >;
  reRenderTable: React.Dispatch<React.SetStateAction<boolean>>;
}

const UserModalComponent = ({
  userDetails,
  open,
  closeFunction,
  reRenderTable,
}: UserModalComponentProps): JSX.Element => {
  const onVerify = (): void => {
    verifyUser(userDetails)
      .then(() => {
        window.alert("User updated");
        closeFunction((currentState) => ({ ...currentState, open: false }));
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
    reRenderTable(true);
  };

  const onDelete = (): void => {
    deleteUser(userDetails)
      .then(() => {
        window.alert("User deleted");
        closeFunction((currentState) => ({ ...currentState, open: false }));
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
    reRenderTable(true);
  };

  return (
    <CModal
      alignment="center"
      visible={open}
      onClose={() =>
        closeFunction((currentState) => ({ ...currentState, open: false }))
      }
    >
      <CModalBody>
        <div className={styles.userDetailsDivStyle}>
          <div className={styles.userDetailsTitleStyle}>User Details</div>
          <CRow>
            <CCol md={3}>Name</CCol>
            <CCol>{userDetails.name}</CCol>
          </CRow>
          <CRow>
            <CCol md={3}>Email</CCol>
            <CCol>{userDetails.email}</CCol>
          </CRow>
          <CRow>
            <CCol md={3}>Birthday</CCol>
            <CCol>
              {moment(userDetails.birthday)
                .utc()
                .format("DD-MMM-YYYY")
                .toString()}
            </CCol>
          </CRow>
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
          <CButton onClick={() => onDelete()}>Delete</CButton>
          <CButton
            onClick={() => onVerify()}
            disabled={userDetails.verified === true}
          >
            Verify
          </CButton>
        </div>
      </CModalBody>
    </CModal>
  );
};

export default UserModalComponent;
