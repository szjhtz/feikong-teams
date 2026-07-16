import { createAsyncThunk } from "@reduxjs/toolkit";
import { listSessions, getSession } from "@/api/sessions";
import { sessionsActions } from "@/app/store";

let sessionsRequest: ReturnType<typeof listSessions> | undefined;

export const loadSessions = createAsyncThunk(
  "sessions/load",
  async (_, { dispatch }) => {
    dispatch(sessionsActions.setSessionsLoading(true));
    sessionsRequest = listSessions();
    try {
      const result = await sessionsRequest;
      dispatch(sessionsActions.setSessions(result.sessions || []));
    } catch (error) {
      dispatch(sessionsActions.setSessionsLoading(false));
      throw error;
    } finally {
      sessionsRequest = undefined;
    }
  },
  { condition: () => sessionsRequest === undefined },
);

export const loadSessionDetail = createAsyncThunk("sessions/detail", async (sessionID: string) => getSession(sessionID));
