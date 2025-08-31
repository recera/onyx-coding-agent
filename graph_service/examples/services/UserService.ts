// UserService.ts - Production code that will be tested

export interface User {
  id: number;
  name: string;
  email: string;
  createdAt?: Date;
  updatedAt?: Date;
}

export interface Database {
  query(sql: string, params?: any[]): Promise<any[]>;
  insert(table: string, data: any): Promise<any>;
  update(table: string, id: number, data: any): Promise<any>;
  delete(table: string, id: number): Promise<boolean>;
}

export class UserService {
  private db: Database;

  constructor(database: Database) {
    this.db = database;
  }

  async getUser(id: number): Promise<User> {
    const results = await this.db.query('SELECT * FROM users WHERE id = ?', [id]);
    
    if (results.length === 0) {
      throw new Error('User not found');
    }
    
    return results[0] as User;
  }

  async createUser(userData: Omit<User, 'id'>): Promise<User> {
    // Validate email
    if (!this.isValidEmail(userData.email)) {
      throw new Error('Invalid email format');
    }

    const created = await this.db.insert('users', userData);
    return created as User;
  }

  async updateUser(id: number, updates: Partial<User>): Promise<User> {
    const updated = await this.db.update('users', id, updates);
    return updated as User;
  }

  async deleteUser(id: number): Promise<boolean> {
    return await this.db.delete('users', id);
  }

  async getUsersByRole(role: string): Promise<User[]> {
    const results = await this.db.query('SELECT * FROM users WHERE role = ?', [role]);
    return results as User[];
  }

  async searchUsers(searchTerm: string): Promise<User[]> {
    const query = `SELECT * FROM users WHERE name LIKE ? OR email LIKE ?`;
    const param = `%${searchTerm}%`;
    const results = await this.db.query(query, [param, param]);
    return results as User[];
  }

  private isValidEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }
}