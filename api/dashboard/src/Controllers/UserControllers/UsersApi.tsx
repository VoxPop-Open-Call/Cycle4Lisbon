import config from "../../config";
import axiosInstance from "../AxiosInterceptor";

export interface UserProps {
  birthday: string;
  createdAt: Date;
  credits: number;
  email: string;
  gender: string;
  id: string;
  initiative: {
    createdAt: Date;
    credits: number;
    description: string;
    enabled: boolean;
    endDate: string;
    goal: number;
    id: string;
    sponsor: {
      createdAt: Date;
      description: string;
      id: string;
      name: string;
      updatedAt: Date;
    };
    sponsorId: string;
    title: string;
    updatedAt: Date;
  };
  initiativeId: string;
  name: string;
  subject: string;
  totalDist: number;
  tripCount: number;
  updatedAt: Date;
  username: string;
  verified: boolean;
  image: UserImageProps;
}

export interface UserImageProps {
  url: string;
  method: string;
}

export const getUserList = async (data: {
  limit: number;
  offset: number;
  orderBy: string;
}): Promise<{
  data: UserProps[];
}> => {
  const response = await axiosInstance.get(`${config.API_URL}/users`, {
    params: {
      limit: data.limit,
      offset: data.offset,
      orderBy: data.orderBy,
    },
  });
  return response;
};

export const verifyUser = async (
  data: UserProps
): Promise<{ data: UserProps }> => {
  const response = await axiosInstance.put(
    `${config.API_URL}/users/${data.id}/verify`
  );
  return response;
};

export const deleteUser = async (
  data: UserProps
): Promise<{ data: UserProps }> => {
  const response = await axiosInstance.delete(
    `${config.API_URL}/users/${data.id}`
  );
  return response;
};

export const getUserDetails = async (
  userId: string
): Promise<{ data: UserProps }> => {
  const response = await axiosInstance.get(`${config.API_URL}/users/${userId}`);
  return response;
};

export const getUserImage = async (
  userId: string
): Promise<{ data: UserImageProps }> => {
  const response = await axiosInstance.get(
    `${config.API_URL}/users/${userId}/picture-get-url`
  );
  return response;
};
