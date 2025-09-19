"use client";

import { createContext, useContext, useState, useEffect, ReactNode } from "react";
import { User, AuthState, AuthContextType } from "@/types";
import { setCookie, getCookie, deleteCookie } from "@/lib/cookies";
import { userAPI, authAPI } from "@/lib/api";

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
	const [authState, setAuthState] = useState<AuthState>({
		isAuthenticated: false,
		user: null,
		token: null,
		loading: true,
	});

	useEffect(() => {
		// Check for existing user info on mount
		const userInfo = getCookie("user_info");

		if (userInfo) {
			try {
				const user = JSON.parse(userInfo);
				setAuthState({
					isAuthenticated: true,
					user,
					token: null, // Token is in HTTP-only cookie, not accessible from client
					loading: false,
				});
			} catch (error) {
				console.error("Error parsing user info:", error);
				logout().catch(console.error);
			}
		} else {
			setAuthState((prev) => ({ ...prev, loading: false }));
		}
	}, []);

	const login = (token: string, user: User) => {
		// Token is set in HTTP-only cookie by backend, not accessible from client
		setCookie("user_info", JSON.stringify(user), 7);
		setAuthState({
			isAuthenticated: true,
			user,
			token: null, // Token is in HTTP-only cookie, not accessible from client
			loading: false,
		});
	};

	const logout = async () => {
		try {
			// Call backend to clear HTTP-only cookie
			await authAPI.logout();
		} catch (error) {
			// Error already handled by toast in API client
			console.error("Logout error:", error);
		}

		// Clear frontend cookies (JWT token is cleared by backend)
		deleteCookie("user_info");

		setAuthState({
			isAuthenticated: false,
			user: null,
			token: null, // Token is in HTTP-only cookie, not accessible from client
			loading: false,
		});

		// Force redirect to login after logout
		window.location.href = "/login";
	};

	const checkAuth = async () => {
		try {
			// Call backend to validate token and get user state
			const response = await userAPI.getState();
			// Map UserState to User type
			const user: User = {
				id: response.data.id,
				name: response.data.name,
				email: response.data.email,
				avatar_url: response.data.avatar_url,
				is_active: response.data.is_active,
				is_deleted: response.data.is_deleted,
				created_at: response.data.created_at,
				profile: response.data.profile,
				room: response.data.room,
				messages: response.data.messages,
			};
			setAuthState({
				isAuthenticated: true,
				user,
				token: null, // Token is in HTTP-only cookie, not accessible from client
				loading: false,
			});
		} catch (error) {
			console.error("Auth check error:", error);
			await logout();
		}
	};

	const value: AuthContextType = {
		...authState,
		login,
		logout,
		checkAuth,
	};

	return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
	const context = useContext(AuthContext);
	if (!context) {
		throw new Error("useAuth must be used within AuthProvider");
	}
	return context;
}
