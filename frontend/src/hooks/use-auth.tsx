"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { userAPI, authAPI } from "@/lib/api";

interface User {
	id: string;
	email: string;
	name: string;
	avatar: string;
	gender: string;
}

interface AuthContextType {
	user: User | null;
	token: string | null;
	isLoading: boolean;
	login: (userData: User) => void;
	logout: () => void;
	checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
	const [user, setUser] = useState<User | null>(null);
	const [token, setToken] = useState<string | null>(null);
	const [isLoading, setIsLoading] = useState(true);

	const login = (userData: User) => {
		setUser(userData);
		setToken(null); // Token is now in HTTP-only cookie
		// Store user info in localStorage for persistence
		localStorage.setItem("user_info", JSON.stringify(userData));
	};

	const logout = async () => {
		try {
			await authAPI.logout();
		} catch (error) {
			console.error("Logout error:", error);
		} finally {
			setUser(null);
			setToken(null);
			localStorage.removeItem("user_info");
		}
	};

	const checkAuth = async () => {
		try {
			const response = await userAPI.getState();
			// Map UserState to User type
			const user: User = {
				id: response.data.id,
				email: response.data.email || "",
				name: response.data.name || "",
				avatar: response.data.avatar_url || "",
				gender:
					response.data.profile?.is_male === true
						? "male"
						: response.data.profile?.is_male === false
						? "female"
						: "other",
			};
			setUser(user);
			setToken(null); // Token is in HTTP-only cookie
		} catch (error) {
			console.error("Auth check failed:", error);
			setUser(null);
			setToken(null);
			localStorage.removeItem("user_info");
		} finally {
			setIsLoading(false);
		}
	};

	useEffect(() => {
		checkAuth();
	}, []);

	return (
		<AuthContext.Provider value={{ user, token, isLoading, login, logout, checkAuth }}>
			{children}
		</AuthContext.Provider>
	);
}

export function useAuth() {
	const context = useContext(AuthContext);
	if (context === undefined) {
		throw new Error("useAuth must be used within an AuthProvider");
	}
	return context;
}
