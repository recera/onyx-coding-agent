// Phase 3: Framework Integration Test File
// This file showcases advanced TypeScript framework integration features

import React, { useState, useEffect, createContext, useContext } from 'react';
import express from 'express';
import axios from 'axios';
import { Injectable, Component, NgModule } from '@angular/core';
import { defineComponent, ref, computed } from 'vue';

// ===== REACT PATTERNS =====

// React Component Props Interface
interface ButtonProps {
  label: string;
  onClick: () => void;
  disabled?: boolean;
  variant?: 'primary' | 'secondary';
  children?: React.ReactNode;
}

// React Functional Component with JSX
const Button: React.FC<ButtonProps> = ({ label, onClick, disabled = false, variant = 'primary', children }) => {
  return (
    <button 
      className={`btn btn-${variant}`}
      onClick={onClick}
      disabled={disabled}
      aria-label={label}
    >
      {children || label}
    </button>
  );
};

// Complex React Component with Hooks and API Calls
interface UserListProps {
  endpoint: string;
  pageSize?: number;
}

const UserList: React.FC<UserListProps> = ({ endpoint, pageSize = 10 }) => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    const fetchUsers = async () => {
      try {
        setLoading(true);
        const response = await fetch(`${endpoint}?limit=${pageSize}`);
        const userData = await response.json();
        setUsers(userData);
      } catch (err) {
        setError('Failed to fetch users');
      } finally {
        setLoading(false);
      }
    };
    
    fetchUsers();
  }, [endpoint, pageSize]);
  
  if (loading) return <div className="spinner">Loading...</div>;
  if (error) return <div className="error">{error}</div>;
  
  return (
    <div className="user-list">
      <h2>Users</h2>
      {users.map(user => (
        <div key={user.id} className="user-card">
          <img src={user.avatar} alt={`${user.name} avatar`} />
          <div className="user-info">
            <h3>{user.name}</h3>
            <p>{user.email}</p>
          </div>
        </div>
      ))}
    </div>
  );
};

// Custom React Hook
const useApi = <T>(url: string) => {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await axios.get<T>(url);
        setData(response.data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };
    
    fetchData();
  }, [url]);
  
  return { data, loading, error };
};

// ===== EXPRESS.JS PATTERNS =====

const app = express();

// Middleware functions
const authMiddleware = (req: express.Request, res: express.Response, next: express.NextFunction) => {
  const token = req.headers.authorization;
  if (!token) {
    return res.status(401).json({ error: 'Unauthorized' });
  }
  next();
};

const loggingMiddleware = (req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.log(`${req.method} ${req.path} - ${new Date().toISOString()}`);
  next();
};

// Apply global middleware
app.use(express.json());
app.use(loggingMiddleware);

// API Endpoints
app.get('/api/users', authMiddleware, async (req, res) => {
  try {
    const limit = parseInt(req.query.limit as string) || 10;
    const users = await getUsersFromDatabase(limit);
    res.json(users);
  } catch (error) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

app.post('/api/users', authMiddleware, async (req, res) => {
  try {
    const userData = req.body;
    const newUser = await createUser(userData);
    res.status(201).json(newUser);
  } catch (error) {
    res.status(400).json({ error: 'Invalid user data' });
  }
});

app.put('/api/users/:id', authMiddleware, async (req, res) => {
  try {
    const userId = req.params.id;
    const updateData = req.body;
    const updatedUser = await updateUser(userId, updateData);
    res.json(updatedUser);
  } catch (error) {
    res.status(404).json({ error: 'User not found' });
  }
});

app.delete('/api/users/:id', authMiddleware, async (req, res) => {
  try {
    const userId = req.params.id;
    await deleteUser(userId);
    res.status(204).send();
  } catch (error) {
    res.status(404).json({ error: 'User not found' });
  }
});

// ===== ANGULAR PATTERNS =====

// Angular Service with Dependency Injection
@Injectable({
  providedIn: 'root'
})
class UserService {
  private apiUrl = '/api/users';
  
  constructor(private http: HttpClient) {}
  
  getUsers(): Observable<User[]> {
    return this.http.get<User[]>(this.apiUrl);
  }
  
  getUserById(id: string): Observable<User> {
    return this.http.get<User>(`${this.apiUrl}/${id}`);
  }
  
  createUser(user: CreateUserRequest): Observable<User> {
    return this.http.post<User>(this.apiUrl, user);
  }
}

// Angular Component
@Component({
  selector: 'app-user-dashboard',
  template: `
    <div class="user-dashboard">
      <h1>User Dashboard</h1>
      <app-user-list [users]="users$ | async"></app-user-list>
    </div>
  `
})
class UserDashboardComponent implements OnInit {
  users$: Observable<User[]>;
  
  constructor(private userService: UserService) {}
  
  ngOnInit() {
    this.users$ = this.userService.getUsers();
  }
}

// ===== VUE 3 PATTERNS =====

// Vue 3 Composition API Component
const UserProfile = defineComponent({
  name: 'UserProfile',
  props: {
    userId: {
      type: String,
      required: true
    }
  },
  setup(props) {
    const user = ref<User | null>(null);
    const loading = ref(true);
    const error = ref<string | null>(null);
    
    const fetchUser = async () => {
      try {
        loading.value = true;
        const response = await fetch(`/api/users/${props.userId}`);
        user.value = await response.json();
      } catch (err) {
        error.value = 'Failed to fetch user';
      } finally {
        loading.value = false;
      }
    };
    
    return {
      user,
      loading,
      error,
      fetchUser
    };
  }
});

// ===== API COMMUNICATION PATTERNS =====

// HTTP Client Service
class ApiClient {
  private baseUrl: string;
  
  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }
  
  async get<T>(endpoint: string): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  }
  
  async post<T>(endpoint: string, data: any): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    return response.json();
  }
}

// Axios-based API Service
class UserApiService {
  private api = axios.create({
    baseURL: '/api',
    timeout: 10000,
  });
  
  async getUsers(page: number = 1, limit: number = 10): Promise<User[]> {
    const response = await this.api.get(`/users?page=${page}&limit=${limit}`);
    return response.data;
  }
}

// GraphQL Patterns
const GET_USERS_QUERY = `
  query GetUsers($limit: Int, $offset: Int) {
    users(limit: $limit, offset: $offset) {
      id
      name
      email
      avatar
    }
  }
`;

class GraphQLClient {
  async query<T>(query: string, variables?: any): Promise<T> {
    const response = await fetch('/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        variables,
      }),
    });
    
    const result = await response.json();
    return result.data;
  }
}

// ===== DATA MODELS =====

interface User {
  id: string;
  name: string;
  email: string;
  avatar: string;
  createdAt: Date;
}

type CreateUserRequest = Omit<User, 'id' | 'createdAt'>;

// Database Model (TypeORM example)
import { Entity, PrimaryGeneratedColumn, Column } from 'typeorm';

@Entity('users')
class UserEntity {
  @PrimaryGeneratedColumn('uuid')
  id: string;
  
  @Column()
  name: string;
  
  @Column({ unique: true })
  email: string;
  
  @Column({ nullable: true })
  avatar: string;
}

// ===== UTILITY FUNCTIONS =====

async function getUsersFromDatabase(limit: number): Promise<User[]> {
  return [];
}

async function createUser(userData: CreateUserRequest): Promise<User> {
  return {} as User;
}

async function updateUser(id: string, userData: any): Promise<User> {
  return {} as User;
}

async function deleteUser(id: string): Promise<void> {
  // Simulate user deletion
}

export {
  Button,
  UserList,
  useApi,
  UserService,
  UserDashboardComponent,
  UserProfile,
  ApiClient,
  UserApiService,
  GraphQLClient,
  User,
  UserEntity
};
