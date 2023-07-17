import config from "../../config";
import axiosInstance from "../AxiosInterceptor";

export interface MetricsProps {
  platform: {
    completedInitiatives: number;
    ongoingInitiatives: number;
    totalCledits: number;
    totalInitiatives: number;
  };
  trips: {
    averageCredits: number;
    averageDist: number;
    total: number;
  };
  users: {
    ageGroups: {
      "18<=age<25": number;
      "25<=age<30": number;
      "30<=age<40": number;
      "40<=age<60": number;
      "60<=age<75": number;
      "age<18": number;
      "age>=75": number;
    };
    aveAge: number;
    genderCount: {
      f: number;
      m: number;
      x: number;
    };
    total: number;
  };
}

export const getMetrics = async (): Promise<{
  data: MetricsProps;
}> => {
  const response = await axiosInstance.get(`${config.API_URL}/metrics`);
  return response;
};
