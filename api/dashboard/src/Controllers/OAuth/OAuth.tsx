import axios from "axios";

import config from "../../config";
import axiosInstance from "../AxiosInterceptor";

interface LoginResponse {
  id_token: string;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export const refreshToken = (): Promise<{
  status: number;
  data: LoginResponse;
}> => {
  const body = {
    grant_type: "refresh_token",
    client_id: config.CLIENT_ID,
    client_secret: config.SECRET,
    refresh_token: localStorage.getItem("refreshToken"),
  };

  return axios.post(`${config.ISSUER_URL}/token`, body, {
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
  });
};

export const userLoginInfo = (credentials: {
  username: string;
  password: string;
}): Promise<{ data: LoginResponse }> => {
  const body = {
    ...credentials,
    grant_type: "password",
    client_id: config.CLIENT_ID,
    client_secret: config.SECRET,
    scope: "openid profile email offline_access",
  };

  return axiosInstance.post(`${config.ISSUER_URL}/token`, body, {
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
  });
};
