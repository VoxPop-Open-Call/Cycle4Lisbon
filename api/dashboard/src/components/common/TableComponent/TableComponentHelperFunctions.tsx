import React from "react";

import styles from "./tableComponent.module.scss";

export const userColumns = [
  {
    key: "name",
    _style: { width: "33%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "email",
    _style: { width: "33%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "verified",
    label: "Status",
    _style: { width: "33%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "actions",
    label: "",
    filter: false,
    sorter: false,
    _style: { width: "1%" },
    _props: { className: styles.headerStyle },
  },
];

export const contentColumns = [
  {
    key: "title",
    _style: { width: "40%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "date",
    _style: { width: "20%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "state",
    _style: { width: "20%" },
    label: "Status",
    _props: { className: styles.headerStyle },
  },
  {
    key: "actions",
    label: "",
    filter: false,
    sorter: false,
    _style: { width: "1%" },
    _props: { className: styles.headerStyle },
  },
];

export const initiativeColumns = [
  {
    key: "title",
    _style: { width: "20%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "institution",
    _style: { width: "25%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "goal",
    _style: { width: "15%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "credits",
    _style: { width: "15%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "sponsors",
    _style: { width: "20%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "enabled",
    label: "Status",
    _style: { width: "5%" },
    _props: { className: styles.headerStyle },
  },
  {
    key: "actions",
    label: "",
    filter: false,
    sorter: false,
    _style: { width: "1%" },
    _props: { className: styles.headerStyle },
  },
];

export const verifiedColumnBadge = (
  userVerifiedStatus: boolean
): JSX.Element => {
  let badgeStyle = styles.notVerifiedStatusBadgeStyle;
  if (userVerifiedStatus) {
    badgeStyle = styles.verifiedStatusBadgeStyle;
  }
  const badgeText = userVerifiedStatus ? "Verified" : "Not Verified";
  return (
    <td>
      <div className={badgeStyle}>{badgeText}</div>
    </td>
  );
};

export const stateColumnBadge = (contentStateStatus: string): JSX.Element => {
  let badgeText,
    badgeStyle = styles.rejectedVerifiedStatusBadgeStyle;
  switch (contentStateStatus) {
    case "approved":
      badgeText = "Approved";
      badgeStyle = styles.verifiedStatusBadgeStyle;
      break;
    case "pending":
      badgeText = "Pending";
      badgeStyle = styles.notVerifiedStatusBadgeStyle;
      break;
    default:
      badgeText = "Rejected";
  }
  return (
    <td>
      <div className={badgeStyle}>{badgeText}</div>
    </td>
  );
};
