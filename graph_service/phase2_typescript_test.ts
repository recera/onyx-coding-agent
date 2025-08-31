// Phase 2 TypeScript Features Demo

// 1. Advanced Generics with Constraints
interface Serializable {
    serialize(): string;
}

type ConstrainedGeneric<T extends Serializable> = {
    data: T;
    serialize(): string;
};

type ConditionalType<T> = T extends string ? string[] : T[];

type MappedType<T> = {
    readonly [K in keyof T]: T[K] | null;
};

// 2. Decorators
function Component(config: any) {
    return function <T extends {new(...args: any[]): {}}>(constructor: T) {
        return class extends constructor {
            config = config;
        };
    };
}

function Injectable() {
    return function <T extends {new(...args: any[]): {}}>(constructor: T) {
        return constructor;
    };
}

// 3. React Component with decorators and generics
@Component({
    selector: 'user-list',
    template: '<div>Users</div>'
})
class UserListComponent<T extends User> extends React.Component<UserProps<T>> {
    @observable users: T[] = [];
    
    render() {
        return <div>{this.props.users.map(user => user.name)}</div>;
    }
}

// 4. Angular Service with dependency injection
@Injectable()
class UserService {
    constructor(private http: HttpClient) {}
    
    getUsers(): Observable<User[]> {
        return this.http.get<User[]>('/api/users');
    }
}

// 5. React Hooks
function useUsers<T extends User>(): [T[], (users: T[]) => void] {
    const [users, setUsers] = useState<T[]>([]);
    return [users, setUsers];
}

function useApi<T>(url: string): ApiResult<T> {
    const [data, setData] = useState<T | null>(null);
    const [loading, setLoading] = useState(false);
    
    useEffect(() => {
        setLoading(true);
        fetch(url).then(response => response.json()).then(setData);
    }, [url]);
    
    return { data, loading };
}

// 6. Type-only imports and re-exports
import type { User, UserProps } from './types';
import { Component as ReactComponent } from 'react';
export type { User, UserProps } from './types';
export { UserService } from './services';

// 7. Namespace declarations
namespace Utils {
    export function formatDate(date: Date): string {
        return date.toISOString();
    }
    
    export namespace String {
        export function capitalize(str: string): string {
            return str.charAt(0).toUpperCase() + str.slice(1);
        }
    }
}

// 8. Module declaration
declare module 'some-library' {
    export interface Config {
        apiUrl: string;
    }
    
    export function initialize(config: Config): void;
}

// 9. Advanced generic constraints with infer
type ReturnType<T extends (...args: any) => any> = T extends (...args: any) => infer R ? R : any;

type Parameters<T extends (...args: any) => any> = T extends (...args: infer P) => any ? P : never;

// 10. Vue component patterns
const UserComponent = defineComponent({
    name: 'UserComponent',
    props: {
        user: {
            type: Object as PropType<User>,
            required: true
        }
    },
    setup(props) {
        const userName = computed(() => props.user.name);
        return { userName };
    }
});

// 11. Express.js endpoint patterns
interface ApiResponse<T> {
    data: T;
    success: boolean;
}

class UserController {
    @Get('/users')
    async getUsers(): Promise<ApiResponse<User[]>> {
        return { data: [], success: true };
    }
    
    @Post('/users')
    async createUser(@Body() user: CreateUserDto): Promise<ApiResponse<User>> {
        return { data: user as User, success: true };
    }
}

// 12. Dynamic imports
async function loadModule() {
    const module = await import('./dynamic-module');
    return module.default;
}

// Interfaces and types
interface User {
    id: number;
    name: string;
    email: string;
}

interface UserProps<T extends User = User> {
    users: T[];
    onUserSelect: (user: T) => void;
}

interface CreateUserDto {
    name: string;
    email: string;
}

interface ApiResult<T> {
    data: T | null;
    loading: boolean;
}

// Decorators for method parameters
function Get(path: string) {
    return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
        // Route decorator logic
    };
}

function Post(path: string) {
    return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
        // Route decorator logic
    };
}

function Body() {
    return function (target: any, propertyKey: string, parameterIndex: number) {
        // Parameter decorator logic
    };
} 