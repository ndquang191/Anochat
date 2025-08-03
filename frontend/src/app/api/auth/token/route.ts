import { NextRequest, NextResponse } from "next/server";

export async function GET(request: NextRequest) {
	try {
		// Get token from HTTP-only cookie
		const token = request.cookies.get("jwt_token")?.value;

		if (!token) {
			return NextResponse.json({ error: "No authentication token found" }, { status: 401 });
		}

		return NextResponse.json({ token });
	} catch (error) {
		console.error("Error getting token:", error);
		return NextResponse.json({ error: "Internal server error" }, { status: 500 });
	}
}
