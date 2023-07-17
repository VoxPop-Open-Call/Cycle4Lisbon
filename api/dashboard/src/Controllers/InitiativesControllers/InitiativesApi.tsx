import config from "../../config";
import axiosInstance from "../AxiosInterceptor";

export interface InitiativesProps {
  createdAt: Date;
  credits: number;
  description: string;
  enabled: boolean;
  endDate: string;
  goal: number;
  id: string;
  institution: {
    createdAt: Date;
    description: string;
    id: string;
    name: string;
    presignedLogoURL: string;
    updatedAt: Date;
  };
  institutionId: string;
  presignedImageURL: string;
  sdgs: [
    {
      code: number;
      description: string;
      imageURI: string;
      title: string;
    }
  ];
  sponsors: [
    {
      createdAt: Date;
      description: string;
      id: string;
      name: string;
      presignedLogoURL: string;
      updatedAt: Date;
    }
  ];
  title: string;
  updatedAt: Date;
}

export const getInitiativesList = async (data: {
  limit: number;
  offset: number;
  orderBy: string;
}): Promise<{ data: InitiativesProps[] }> => {
  const response = await axiosInstance.get(`${config.API_URL}/initiatives`, {
    params: {
      limit: data.limit,
      offset: data.offset,
      orderBy: data.orderBy,
      includeDisabled: true,
    },
  });
  return response;
};

export const getInitiativeDetails = async (
  initiativeId: string
): Promise<{ data: InitiativesProps }> => {
  const response = await axiosInstance.get(
    `${config.API_URL}/initiatives/${initiativeId}`
  );
  return response;
};
