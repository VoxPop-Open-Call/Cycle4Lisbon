import React, { useState } from "react";

import { cilEnvelopeClosed, cilLockLocked } from "@coreui/icons";
import CIcon from "@coreui/icons-react";
import {
  CCard,
  CCardBody,
  CFormInput,
  CFormLabel,
  CInputGroup,
  CInputGroupText,
} from "@coreui/react";
import { useNavigate } from "react-router";

import LogoIcon from "../../../assets/loginBackground/logo-cycle4lisbon.svg";
import { userLoginInfo } from "../../../Controllers/OAuth/OAuth";

import styles from "./login.module.scss";

interface LoginProps {
  onLoggedChange: (value: boolean) => void;
}

const Login: React.FC<LoginProps> = ({ onLoggedChange }) => {
  const navigate = useNavigate();
  const [credentials, setCredentials] = useState({
    username: "",
    password: "",
    error: "",
  });

  const handleOnChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
    const { name, value }: { name: string; value: string } = event.target;
    setCredentials((currentState) => ({
      ...currentState,
      [name]: value,
      error: "",
    }));
  };

  const validateCredentials = (data: {
    username: string;
    password: string;
  }): boolean => {
    if (data?.username && data?.password) {
      return true;
    }
    return false;
  };

  const handleSubmit = (): void => {
    if (validateCredentials(credentials)) {
      userLoginInfo({
        username: credentials.username,
        password: credentials.password,
      })
        .then(({ data }) => {
          localStorage.setItem("encodedToken", data.access_token);
          localStorage.setItem("refreshToken", data.refresh_token);
          onLoggedChange(true);
          navigate("/dashboard");
        })
        .catch(({ response }) => {
          setCredentials((state) => ({
            ...state,
            error: response.data.error_description,
          }));
        });
    } else {
      setCredentials((state) => ({
        ...state,
        error: "Both fields need to be filled.",
      }));
    }
  };

  const renderError = (): string | null => {
    if (credentials?.error) {
      return credentials.error;
    }
    return null;
  };

  return (
    <>
      <div
        className={styles.positionDiv}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            handleSubmit();
          }
        }}
      >
        <div className={styles.logoDiv}>
          <img src={LogoIcon} />
        </div>
        <CCard className={styles.cardStyle}>
          <CCardBody className={styles.cardBodyStyle}>
            <div className={styles.titleStyle}>Login</div>
            <div className={styles.inputFieldsDiv}>
              <div className={styles.inputContainerDiv}>
                <CFormLabel className={styles.loginFieldStyle}>
                  Email
                </CFormLabel>
                <CInputGroup className={styles.inputFieldsContainerDiv}>
                  <CInputGroupText className={styles.prependIconStyle}>
                    <CIcon icon={cilEnvelopeClosed} />
                  </CInputGroupText>
                  <CFormInput
                    className={styles.loginFields}
                    name="username"
                    placeholder="example@email.com"
                    autoComplete="username"
                    onChange={(e) => handleOnChange(e)}
                  />
                </CInputGroup>
              </div>

              <div className={styles.inputContainerDiv}>
                <CFormLabel className={styles.loginFieldStyle}>
                  Password
                </CFormLabel>
                <CInputGroup className={styles.inputFieldsContainerDiv}>
                  <CInputGroupText className={styles.prependIconStyle}>
                    <CIcon icon={cilLockLocked} />
                  </CInputGroupText>
                  <CFormInput
                    className={styles.loginFields}
                    name="password"
                    type="password"
                    placeholder="Password"
                    autoComplete="current-password"
                    onChange={(e) => handleOnChange(e)}
                  />
                </CInputGroup>
              </div>
              <div className={styles.errorTextStyle}>{renderError()}</div>
              <button className={styles.buttonStyle} onClick={handleSubmit}>
                Continue
              </button>
            </div>
          </CCardBody>
        </CCard>
      </div>
    </>
  );
};

export default Login;
