import React from "react";

import { cilClock, cilEuro, cilGraph } from "@coreui/icons";
import CIcon from "@coreui/icons-react";

import UserErrorImage from "../../../assets/userIcon/placeholder@2x.png";
import { UserProps } from "../../../Controllers/UserControllers/UsersApi";

import styles from "./userInfoCardComponent.module.scss";

interface UserInfoCardComponentProps {
  userDetails: UserProps;
}

const UserInfoCardComponent: React.FC<UserInfoCardComponentProps> = ({
  userDetails,
}): JSX.Element => {
  const { image, name, username, email, tripCount, totalDist, credits } =
    userDetails;
  return (
    <div>
      <div className={styles.headerContainerDiv}>
        <img
          src={image.url}
          height={121}
          width={121}
          className={styles.userImageStyle}
          onError={(e) => {
            const target = e.target as HTMLImageElement;
            target.src = UserErrorImage;
          }}
        />
        <div>
          <div className={styles.userNameContainerDiv}>
            <div className={styles.nameStyle}>{name}</div>
            <div className={styles.userNameBadgeStyle}>{username}</div>
          </div>
          <div className={styles.emailStyle}>{email}</div>
        </div>
        <div className={styles.userStatsContainerDiv}>
          <div className={styles.singleUserStatContainerDiv}>
            <div>
              <CIcon icon={cilClock} className={styles.userStatsIconStyle} />
            </div>
            <div className={styles.userStatsValueStyle}>
              {Number(tripCount).toFixed(0)}
            </div>
            <div className={styles.userStatsDescriptionStyle}>Total Rides</div>
          </div>
          <div className={styles.separator} />
          <div className={styles.singleUserStatContainerDiv}>
            <div>
              <CIcon icon={cilGraph} className={styles.userStatsIconStyle} />
            </div>
            <div className={styles.userStatsValueStyle}>
              {Number(totalDist).toFixed(2)}km
            </div>
            <div className={styles.userStatsDescriptionStyle}>
              Total Distance
            </div>
          </div>
          <div className={styles.separator} />
          <div className={styles.singleUserStatContainerDiv}>
            <div>
              <CIcon icon={cilEuro} className={styles.userStatsIconStyle} />
            </div>
            <div className={styles.userStatsValueStyle}>
              {Number(credits).toFixed(0)}
            </div>
            <div className={styles.userStatsDescriptionStyle}>
              Total Donations
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default UserInfoCardComponent;
