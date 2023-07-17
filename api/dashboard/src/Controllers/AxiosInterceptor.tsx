import axios from "axios";

import { refreshToken } from "./OAuth/OAuth";

const axiosInstance = axios.create();

const isExpired = (token: string): boolean => {
  const base64Url = token.split(".")[1];
  const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
  const jsonPayload = decodeURIComponent(
    window
      .atob(base64)
      .split("")
      .map(function (c) {
        return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
      })
      .join("")
  );
  const jwtContent = JSON.parse(jsonPayload);
  const expiryDate = jwtContent.exp;
  const currentDate = Date.now() / 1000;
  return currentDate > expiryDate;
};

axiosInstance.interceptors.request.use(async (config) => {
  let token = localStorage.getItem("encodedToken");
  if (token) {
    if (isExpired(token)) {
      try {
        const rs = await refreshToken();
        const { data } = rs;
        localStorage.setItem("encodedToken", data.access_token);
        localStorage.setItem("refreshToken", data.refresh_token);
        token = data.access_token;
      } catch (error: unknown) {
        localStorage.clear();
        window.location.assign("/login");
      }
    }
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default axiosInstance;
