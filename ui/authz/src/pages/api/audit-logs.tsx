import { NextApiRequest, NextApiResponse } from "next";
// import { getSession } from 'next-auth/react';
import axios from "axios";

// This is an API route for fetching authorization audit logs
export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  // const session = await getSession({ req });

  // Ensure the user is authenticated
  // if (!session) {
  //   return res.status(401).json({ error: 'Authentication required' });
  // }

  try {
    // Extract query parameters
    const {
      action_type,
      entity_type,
      entity_id,
      subject_type,
      subject_id,
      result,
      start_time,
      end_time,
      limit = 50,
      offset = 0,
    } = req.query;

    // Build query string
    const queryParams = new URLSearchParams();
    if (action_type) queryParams.append("action_type", String(action_type));
    if (entity_type) queryParams.append("entity_type", String(entity_type));
    if (entity_id) queryParams.append("entity_id", String(entity_id));
    if (subject_type) queryParams.append("subject_type", String(subject_type));
    if (subject_id) queryParams.append("subject_id", String(subject_id));
    if (result) queryParams.append("result", String(result));
    if (start_time) queryParams.append("start_time", String(start_time));
    if (end_time) queryParams.append("end_time", String(end_time));
    if (limit) queryParams.append("limit", String(limit));
    if (offset) queryParams.append("offset", String(offset));

    // Call the backend API
    const response = await axios.get(
      `${
        process.env.NEXT_PUBLIC_API_URL
      }/api/audit/logs?${queryParams.toString()}`,
      {
        headers: {
          // Authorization: `Bearer ${session.accessToken}`,
        },
      }
    );

    // Return the data from the backend
    return res.status(200).json(response.data);
  } catch (error: any) {
    console.error("Error fetching audit logs:", error);
    return res.status(error.response?.status || 500).json({
      error: error.response?.data?.error || "Failed to fetch audit logs",
    });
  }
}
