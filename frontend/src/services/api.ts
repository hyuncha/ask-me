import axios, { AxiosInstance, AxiosError } from 'axios';
import { ApiResponse } from '../types';

const API_BASE_URL = process.env.REACT_APP_API_URL || '';

class ApiService {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
      withCredentials: false,
    });

    // Request interceptor for adding auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('access_token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor for handling errors
    this.client.interceptors.response.use(
      (response) => response,
      async (error: AxiosError) => {
        if (error.response?.status === 401) {
          // Handle token refresh or redirect to login
          localStorage.removeItem('access_token');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  async get<T>(url: string, params?: any): Promise<ApiResponse<T>> {
    try {
      const response = await this.client.get<T>(url, { params });
      return { data: response.data };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async post<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    try {
      const response = await this.client.post<T>(url, data);
      return { data: response.data };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async put<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    try {
      const response = await this.client.put<T>(url, data);
      return { data: response.data };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async delete<T>(url: string): Promise<ApiResponse<T>> {
    try {
      const response = await this.client.delete<T>(url);
      return { data: response.data };
    } catch (error) {
      return this.handleError(error);
    }
  }

  private handleError(error: any): ApiResponse<never> {
    if (axios.isAxiosError(error) && error.response) {
      return {
        error: {
          code: error.response.data?.code || 'UNKNOWN_ERROR',
          message: error.response.data?.message || 'An error occurred',
        },
      };
    }
    return {
      error: {
        code: 'NETWORK_ERROR',
        message: 'Network error occurred',
      },
    };
  }
}

export default new ApiService();
