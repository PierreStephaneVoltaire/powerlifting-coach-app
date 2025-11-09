import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/store/authStore';
import { apiClient } from '@/utils/api';
import { LoginRequest, RegisterRequest } from '@/types';

export const useAuth = () => {
  const { user, tokens, isAuthenticated, login, logout, setLoading } = useAuthStore();
  const queryClient = useQueryClient();

  const loginMutation = useMutation({
    mutationFn: async ({ email, password }: LoginRequest) => {
      setLoading(true);
      const tokens = await apiClient.login(email, password);

      const userResponse = await apiClient.getProfile();
      
      return { tokens, userResponse };
    },
    onSuccess: ({ tokens, userResponse }) => {
      login(tokens, userResponse);
      queryClient.invalidateQueries({ queryKey: ['user'] });
    },
    onError: () => {
      setLoading(false);
    },
  });

  const registerMutation = useMutation({
    mutationFn: async ({ email, password, name, user_type }: RegisterRequest) => {
      setLoading(true);
      const tokens = await apiClient.register(email, password, name, user_type);

      const userResponse = await apiClient.getProfile();
      
      return { tokens, userResponse };
    },
    onSuccess: ({ tokens, userResponse }) => {
      login(tokens, userResponse);
      queryClient.invalidateQueries({ queryKey: ['user'] });
    },
    onError: () => {
      setLoading(false);
    },
  });

  const logoutMutation = useMutation({
    mutationFn: async () => {
      if (tokens?.refresh_token) {
        await apiClient.logout(tokens.refresh_token);
      }
    },
    onSettled: () => {
      logout();
      queryClient.clear();
    },
  });

  const userQuery = useQuery({
    queryKey: ['user'],
    queryFn: () => apiClient.getProfile(),
    enabled: isAuthenticated,
    retry: false,
  });

  return {
    user,
    tokens,
    isAuthenticated,
    isLoading: useAuthStore((state) => state.isLoading) || loginMutation.isPending || registerMutation.isPending,
    login: loginMutation.mutateAsync,
    register: registerMutation.mutateAsync,
    logout: logoutMutation.mutate,
    userQuery,
    loginError: loginMutation.error,
    registerError: registerMutation.error,
  };
};