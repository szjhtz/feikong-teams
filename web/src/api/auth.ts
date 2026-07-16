import { post } from "./client";

export interface LoginResponse {
  authenticated: boolean;
}

export function login(username: string, password: string) {
	return post<LoginResponse>("/api/fkteams/login", { username, password, cookie_only: true }, { authFailure: "ignore" });
}
