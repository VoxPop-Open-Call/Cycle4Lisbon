import React, { useCallback, useMemo, useState } from "react";

import { useParams } from "react-router";

import {
  InitiativesProps,
  getInitiativeDetails,
} from "../../../Controllers/InitiativesControllers/InitiativesApi";

import InitiativesDetails from "./InitiativesDetails";

const InitiativesDetailsManager: React.FC = (): JSX.Element => {
  const params = useParams();
  const [initiativeDetails, setInitiativeDetails] = useState<InitiativesProps>(
    {} as InitiativesProps
  );

  const getInitiativeDetailsFetch = useCallback(() => {
    getInitiativeDetails(params.id as string)
      .then(({ data }) => {
        setInitiativeDetails(data);
      })
      .catch(({ response }) => {
        window.alert(response.data.error.message);
      });
  }, [params.id]);

  useMemo(() => {
    getInitiativeDetailsFetch();
  }, [getInitiativeDetailsFetch]);

  const renderInitiativeDetails = (): JSX.Element => {
    if (!Object.keys(initiativeDetails).length) {
      return <></>;
    }
    return <InitiativesDetails details={initiativeDetails} />;
  };

  return renderInitiativeDetails();
};

export default InitiativesDetailsManager;
