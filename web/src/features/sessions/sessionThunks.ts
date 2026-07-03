import { createAsyncThunk } from "@reduxjs/toolkit";
import { listSessions, getSession } from "@/api/sessions";
import { sessionsActions } from "@/app/store";

export const loadSessions = createAsyncThunk("sessions/load", async (_, { dispatch }) => {
  dispatch(sessionsActions.setSessionsLoading(true));
  const result = await listSessions();
  dispatch(sessionsActions.setSessions(result.sessions || []));
});

export const loadSessionDetail = createAsyncThunk("sessions/detail", async (sessionID: string) => getSession(sessionID));
