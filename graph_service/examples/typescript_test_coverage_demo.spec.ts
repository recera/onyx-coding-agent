// Comprehensive TypeScript Test Coverage Demo
// This file demonstrates all supported test patterns for the TypeScript analyzer

import { describe, it, test, expect, beforeEach, afterEach, beforeAll, afterAll } from '@jest/globals';
import { jest } from '@jest/globals';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { mount, shallow } from 'enzyme';
import * as sinon from 'sinon';
import { vi } from 'vitest';

// Import production code to test
import { UserService } from './services/UserService';
import { calculateTotal, processOrder, validateEmail } from './utils/helpers';
import { Button } from './components/Button';
import { UserList } from './components/UserList';
import { api } from './api/client';

// Test Suite with nested suites
describe('UserService Test Suite', () => {
  let userService: UserService;
  let mockDatabase: any;
  let consoleSpy: sinon.SinonSpy;

  // Test hooks
  beforeAll(() => {
    console.log('Starting UserService tests');
  });

  afterAll(() => {
    console.log('Finished UserService tests');
  });

  beforeEach(() => {
    // Setup mock database
    mockDatabase = {
      query: jest.fn(),
      insert: jest.fn(),
      update: jest.fn(),
      delete: jest.fn()
    };
    userService = new UserService(mockDatabase);
    consoleSpy = sinon.spy(console, 'log');
  });

  afterEach(() => {
    jest.clearAllMocks();
    consoleSpy.restore();
  });

  // Nested test suite
  describe('User CRUD Operations', () => {
    describe('getUser', () => {
      it('should fetch user by id successfully', async () => {
        // Arrange
        const mockUser = { id: 1, name: 'John Doe', email: 'john@example.com' };
        mockDatabase.query.mockResolvedValue([mockUser]);

        // Act
        const user = await userService.getUser(1);

        // Assert
        expect(user).toEqual(mockUser);
        expect(mockDatabase.query).toHaveBeenCalledWith('SELECT * FROM users WHERE id = ?', [1]);
        expect(mockDatabase.query).toHaveBeenCalledTimes(1);
      });

      it('should throw error when user not found', async () => {
        mockDatabase.query.mockResolvedValue([]);

        await expect(userService.getUser(999)).rejects.toThrow('User not found');
        expect(mockDatabase.query).toHaveBeenCalled();
      });
    });

    describe('createUser', () => {
      test('should create a new user with valid data', async () => {
        const newUser = { name: 'Jane Doe', email: 'jane@example.com' };
        const createdUser = { id: 2, ...newUser };
        mockDatabase.insert.mockResolvedValue(createdUser);

        const result = await userService.createUser(newUser);

        expect(result).toEqual(createdUser);
        expect(mockDatabase.insert).toHaveBeenCalledWith('users', newUser);
        expect(result.id).toBeDefined();
      });

      test.skip('should validate email format', async () => {
        // This test is skipped for now
        const invalidUser = { name: 'Invalid', email: 'not-an-email' };
        await expect(userService.createUser(invalidUser)).rejects.toThrow('Invalid email');
      });
    });

    describe('updateUser', () => {
      it.only('should update existing user', async () => {
        // This test runs exclusively
        const updates = { name: 'Updated Name' };
        mockDatabase.update.mockResolvedValue({ id: 1, ...updates });

        const result = await userService.updateUser(1, updates);

        expect(result.name).toBe('Updated Name');
        expect(mockDatabase.update).toHaveBeenCalledWith('users', 1, updates);
      });
    });
  });

  // Parameterized tests
  describe.each([
    [1, 'John'],
    [2, 'Jane'],
    [3, 'Bob'],
  ])('User %i operations', (userId, userName) => {
    test(`should handle user ${userName} correctly`, async () => {
      const mockUser = { id: userId, name: userName };
      mockDatabase.query.mockResolvedValue([mockUser]);

      const user = await userService.getUser(userId);
      
      expect(user.name).toBe(userName);
      expect(user.id).toBe(userId);
    });
  });
});

// React Component Tests
describe('Button Component Tests', () => {
  describe('Rendering', () => {
    it('should render button with correct label', () => {
      render(<Button label="Click me" />);
      
      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
      expect(button).toHaveTextContent('Click me');
    });

    test('should apply custom className', () => {
      const { container } = render(<Button label="Test" className="custom-class" />);
      
      const button = container.querySelector('.custom-class');
      expect(button).toBeTruthy();
    });
  });

  describe('Interactions', () => {
    it('should handle click events', async () => {
      const handleClick = jest.fn();
      render(<Button label="Clickable" onClick={handleClick} />);
      
      const button = screen.getByRole('button');
      fireEvent.click(button);
      
      await waitFor(() => {
        expect(handleClick).toHaveBeenCalledTimes(1);
      });
    });

    it('should be disabled when disabled prop is true', () => {
      render(<Button label="Disabled" disabled={true} />);
      
      const button = screen.getByRole('button') as HTMLButtonElement;
      expect(button.disabled).toBe(true);
      expect(button).toHaveAttribute('disabled');
    });
  });

  // Snapshot testing
  describe('Snapshots', () => {
    it('should match snapshot', () => {
      const component = render(<Button label="Snapshot Test" />);
      expect(component).toMatchSnapshot();
    });

    test('should match inline snapshot', () => {
      const html = render(<Button label="Inline" />).container.innerHTML;
      expect(html).toMatchInlineSnapshot(`"<button>Inline</button>"`);
    });
  });
});

// Enzyme Component Tests
describe('UserList Component (Enzyme)', () => {
  it('should shallow render without errors', () => {
    const wrapper = shallow(<UserList users={[]} />);
    expect(wrapper.exists()).toBe(true);
  });

  test('should mount and display users', () => {
    const users = [
      { id: 1, name: 'User 1' },
      { id: 2, name: 'User 2' }
    ];
    
    const wrapper = mount(<UserList users={users} />);
    
    expect(wrapper.find('li')).toHaveLength(2);
    expect(wrapper.text()).toContain('User 1');
    expect(wrapper.text()).toContain('User 2');
  });
});

// API Integration Tests
describe('API Integration Tests', () => {
  describe('User API', () => {
    it('should fetch users from API', async () => {
      const response = await fetch('/api/users');
      const users = await response.json();
      
      expect(response.status).toBe(200);
      expect(Array.isArray(users)).toBe(true);
    });

    test('should create user via API', async () => {
      const newUser = { name: 'API User', email: 'api@example.com' };
      
      const response = await fetch('/api/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newUser)
      });
      
      expect(response.status).toBe(201);
      const created = await response.json();
      expect(created).toHaveProperty('id');
    });
  });

  describe('Order API', () => {
    it('should process order through API', async () => {
      const order = { items: ['item1', 'item2'], total: 100 };
      
      const response = await api.post('/api/orders', order);
      
      expect(response.status).toBe(200);
      expect(response.data).toHaveProperty('orderId');
    });
  });
});

// Utility Function Tests
describe('Helper Functions', () => {
  describe('calculateTotal', () => {
    test('should calculate sum correctly', () => {
      expect(calculateTotal([10, 20, 30])).toBe(60);
      expect(calculateTotal([])).toBe(0);
      expect(calculateTotal([5])).toBe(5);
    });

    it('should handle negative numbers', () => {
      expect(calculateTotal([-10, 20, -5])).toBe(5);
    });
  });

  describe('validateEmail', () => {
    test.each([
      ['valid@example.com', true],
      ['invalid.email', false],
      ['test@domain.co.uk', true],
      ['no-at-sign.com', false],
    ])('validateEmail(%s) should return %s', (email, expected) => {
      expect(validateEmail(email)).toBe(expected);
    });
  });

  describe('processOrder', () => {
    it('should process order with mocked dependencies', async () => {
      const orderSpy = jest.spyOn(api, 'post').mockResolvedValue({ data: { success: true } });
      
      const result = await processOrder({ items: ['item1'] });
      
      expect(result.success).toBe(true);
      expect(orderSpy).toHaveBeenCalledWith(expect.stringContaining('/orders'), expect.any(Object));
      
      orderSpy.mockRestore();
    });
  });
});

// Async Tests
describe('Async Operations', () => {
  it('should handle promises correctly', async () => {
    const promise = Promise.resolve('success');
    await expect(promise).resolves.toBe('success');
  });

  test('should handle promise rejection', async () => {
    const promise = Promise.reject(new Error('failed'));
    await expect(promise).rejects.toThrow('failed');
  });

  it('should work with async/await', async () => {
    const result = await userService.getUser(1);
    expect(result).toBeDefined();
  });
});

// Mock and Spy Examples
describe('Mocking and Spying', () => {
  describe('Jest Mocks', () => {
    it('should use jest.fn() for function mocks', () => {
      const mockFn = jest.fn();
      mockFn('arg1', 'arg2');
      
      expect(mockFn).toHaveBeenCalledWith('arg1', 'arg2');
      expect(mockFn.mock.calls).toHaveLength(1);
    });

    test('should mock modules', () => {
      jest.mock('./api/client');
      // Module is now mocked
    });

    it('should spy on methods', () => {
      const obj = { method: () => 'original' };
      const spy = jest.spyOn(obj, 'method').mockReturnValue('mocked');
      
      expect(obj.method()).toBe('mocked');
      expect(spy).toHaveBeenCalled();
      
      spy.mockRestore();
    });
  });

  describe('Sinon Mocks', () => {
    it('should create sinon stubs', () => {
      const stub = sinon.stub().returns('stubbed');
      
      expect(stub()).toBe('stubbed');
      expect(stub.calledOnce).toBe(true);
    });

    test('should create sinon spies', () => {
      const obj = { method: () => 'original' };
      const spy = sinon.spy(obj, 'method');
      
      obj.method();
      
      expect(spy.calledOnce).toBe(true);
      spy.restore();
    });
  });

  describe('Vitest Mocks', () => {
    it('should use vi.fn() for mocks', () => {
      const mock = vi.fn();
      mock('test');
      
      expect(mock).toHaveBeenCalledWith('test');
    });
  });
});

// Test Fixtures
describe('Test Fixtures', () => {
  const userFixture = {
    id: 1,
    name: 'Test User',
    email: 'test@example.com',
    role: 'admin'
  };

  const ordersFixture = [
    { id: 1, total: 100 },
    { id: 2, total: 200 }
  ];

  it('should use user fixture', () => {
    expect(userFixture.name).toBe('Test User');
    expect(userFixture.role).toBe('admin');
  });

  test('should use orders fixture', () => {
    expect(ordersFixture).toHaveLength(2);
    expect(ordersFixture[0].total).toBe(100);
  });
});

// Error Handling Tests
describe('Error Handling', () => {
  it('should throw specific errors', () => {
    expect(() => {
      throw new Error('Custom error');
    }).toThrow('Custom error');
  });

  test('should not throw for valid input', () => {
    expect(() => {
      validateEmail('valid@email.com');
    }).not.toThrow();
  });
});

// Performance Tests
describe('Performance Tests', () => {
  it('should complete within time limit', async () => {
    const start = Date.now();
    
    await userService.getUser(1);
    
    const duration = Date.now() - start;
    expect(duration).toBeLessThan(100); // Should complete within 100ms
  });
});

// Suite with context
context('Alternative Suite Syntax', () => {
  specify('should support context/specify syntax', () => {
    expect(true).toBe(true);
  });
});

// Mocha-style suite
suite('Mocha Style Tests', function() {
  test('should support suite/test syntax', () => {
    expect(1 + 1).toBe(2);
  });
});