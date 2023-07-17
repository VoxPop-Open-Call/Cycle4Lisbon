export const getHttpHeaders = (
  method: string
): Record<string, string | object> => {
  const init: Record<string, string | object> = {};
  init.method = method;
  const headers: Record<string, string> = {};
  headers.Accept = "*/*";
  headers["Content-type"] = "application/json";
  const token = localStorage.getItem("encodedToken");
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  init.headers = headers;
  return init;
};

export const httpGetRequest = async (
  callbackMethod: () => Promise<Response>
): Promise<object> => {
  const response = await callbackMethod();
  return response.json();
};

export const httpPostRequest = async (
  callbackMethod: () => Promise<Response>
): Promise<object> => {
  const response = await callbackMethod();
  return response.json();
};
