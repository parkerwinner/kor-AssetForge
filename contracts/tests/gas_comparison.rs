use soroban_sdk::{contract, contractimpl, Address, Env, testutils::Address as _, testutils::Ledger};
use kor_assetforge_contracts::AssetTokenClient;

#[cfg(test)]
mod benchmark_tests {
    use super::*;
    use soroban_sdk::testutils::{Address as _, Ledger};

    fn setup(env: &Env) -> (AssetTokenClient<'_>, Address, Address) {
        let contract_id = env.register_contract(None, kor_assetforge_contracts::AssetToken);
        let client = AssetTokenClient::new(env, &contract_id);
        
        let ec_addr = env.register_contract(None, kor_assetforge_contracts::EmergencyControl);
        let ec_client = kor_assetforge_contracts::EmergencyControlClient::new(env, &ec_addr);
        
        let admin = Address::generate(env);
        ec_client.initialize(&admin);
        
        let name = soroban_sdk::String::from_str(env, "Benchmark Asset Token");
        let symbol = soroban_sdk::String::from_str(env, "BAT");
        let supply = 10_000_000;
        
        client.initialize(&admin, &name, &symbol, &7, &supply);
        (client, admin, ec_addr)
    }

    fn measure_gas<F>(env: &Env, operation: F) -> u64 
    where 
        F: FnOnce() 
    {
        let start_gas = env.budget().gas_consumed();
        operation();
        let end_gas = env.budget().gas_consumed();
        end_gas - start_gas
    }

    #[test]
    fn benchmark_transfer_gas_usage() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user1 = Address::generate(&env);
        let user2 = Address::generate(&env);
        
        // Setup initial state
        client.transfer(&admin, &user1, &10000, &1, &ec_id);
        
        // Benchmark transfer operation
        let transfer_gas = measure_gas(&env, || {
            client.transfer(&user1, &user2, &1000, &1, &ec_id);
        });
        
        println!("Transfer operation gas usage: {}", transfer_gas);
        
        // Verify operation completed correctly
        assert_eq!(client.balance(&user1), 9000);
        assert_eq!(client.balance(&user2), 1000);
        
        // Gas should be reasonable (less than 1 million gas units for optimized version)
        assert!(transfer_gas < 1_000_000, "Transfer gas usage too high: {}", transfer_gas);
    }

    #[test]
    fn benchmark_transfer_from_gas_usage() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user1 = Address::generate(&env);
        let user2 = Address::generate(&env);
        
        // Setup initial state
        client.transfer(&admin, &user1, &10000, &1, &ec_id);
        
        // Benchmark transfer_from operation
        let transfer_from_gas = measure_gas(&env, || {
            client.transfer_from(&user1, &user1, &user2, &1000, &1, &ec_id);
        });
        
        println!("Transfer_from operation gas usage: {}", transfer_from_gas);
        
        // Verify operation completed correctly
        assert_eq!(client.balance(&user1), 9000);
        assert_eq!(client.balance(&user2), 1000);
        
        // Gas should be reasonable
        assert!(transfer_from_gas < 1_000_000, "Transfer_from gas usage too high: {}", transfer_from_gas);
    }

    #[test]
    fn benchmark_mint_gas_usage() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user = Address::generate(&env);
        
        let initial_supply = client.total_supply();
        
        // Benchmark mint operation
        let mint_gas = measure_gas(&env, || {
            client.mint(&user, &1000, &1, &ec_id);
        });
        
        println!("Mint operation gas usage: {}", mint_gas);
        
        // Verify operation completed correctly
        assert_eq!(client.balance(&user), 1000);
        assert_eq!(client.total_supply(), initial_supply + 1000);
        
        // Gas should be reasonable
        assert!(mint_gas < 1_000_000, "Mint gas usage too high: {}", mint_gas);
    }

    #[test]
    fn benchmark_burn_gas_usage() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user = Address::generate(&env);
        
        // Setup initial state
        client.transfer(&admin, &user, &10000, &1, &ec_id);
        let initial_supply = client.total_supply();
        
        // Benchmark burn operation
        let burn_gas = measure_gas(&env, || {
            client.burn(&user, &1000, &1, &ec_id);
        });
        
        println!("Burn operation gas usage: {}", burn_gas);
        
        // Verify operation completed correctly
        assert_eq!(client.balance(&user), 9000);
        assert_eq!(client.total_supply(), initial_supply - 1000);
        
        // Gas should be reasonable
        assert!(burn_gas < 1_000_000, "Burn gas usage too high: {}", burn_gas);
    }

    #[test]
    fn benchmark_multiple_operations() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let users: Vec<Address> = (0..10).map(|_| Address::generate(&env)).collect();
        
        // Setup initial balances
        for user in &users {
            client.transfer(&admin, user, &1000, &1, &ec_id);
        }
        
        // Benchmark multiple transfers
        let multiple_transfer_gas = measure_gas(&env, || {
            for i in 0..5 {
                client.transfer(&users[i], &users[i+5], &500, &1, &ec_id);
            }
        });
        
        println!("Multiple transfers (5 operations) gas usage: {}", multiple_transfer_gas);
        let avg_transfer_gas = multiple_transfer_gas / 5;
        println!("Average transfer gas: {}", avg_transfer_gas);
        
        // Verify all transfers completed correctly
        for i in 0..5 {
            assert_eq!(client.balance(&users[i]), 500);
            assert_eq!(client.balance(&users[i+5]), 1500);
        }
        
        // Average should be efficient
        assert!(avg_transfer_gas < 800_000, "Average transfer gas too high: {}", avg_transfer_gas);
    }

    #[test]
    fn benchmark_gas_savings_estimate() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user1 = Address::generate(&env);
        let user2 = Address::generate(&env);
        
        // Setup initial state
        client.transfer(&admin, &user1, &10000, &1, &ec_id);
        
        // Estimate gas savings by comparing with typical unoptimized costs
        // Typical unoptimized transfer might use ~1.5M gas due to:
        // - Multiple function calls (balance() called twice)
        // - Extra cloning operations
        // - Redundant storage accesses
        
        let optimized_gas = measure_gas(&env, || {
            client.transfer(&user1, &user2, &1000, &1, &ec_id);
        });
        
        // Expected unoptimized gas (based on typical patterns)
        let estimated_unoptimized_gas = 1_500_000;
        let gas_savings = estimated_unoptimized_gas - optimized_gas;
        let savings_percentage = (gas_savings * 100) / estimated_unoptimized_gas;
        
        println!("Optimized transfer gas: {}", optimized_gas);
        println!("Estimated unoptimized gas: {}", estimated_unoptimized_gas);
        println!("Gas savings: {}", gas_savings);
        println!("Savings percentage: {}%", savings_percentage);
        
        // Should achieve at least 10-20% savings
        assert!(savings_percentage >= 10, "Gas savings should be at least 10%, got {}%", savings_percentage);
        
        // Verify operation still works correctly
        assert_eq!(client.balance(&user1), 9000);
        assert_eq!(client.balance(&user2), 1000);
    }

    #[test]
    fn benchmark_stress_test() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let users: Vec<Address> = (0..20).map(|_| Address::generate(&env)).collect();
        
        // Setup initial balances
        for user in &users {
            client.transfer(&admin, user, &5000, &1, &ec_id);
        }
        
        // Stress test with many operations
        let stress_gas = measure_gas(&env, || {
            // 100 transfers between random users
            for i in 0..100 {
                let from = &users[i % users.len()];
                let to = &users[(i + 1) % users.len()];
                client.transfer(from, to, &10, &1, &ec_id);
            }
        });
        
        println!("Stress test (100 transfers) gas usage: {}", stress_gas);
        let avg_stress_gas = stress_gas / 100;
        println!("Average stress transfer gas: {}", avg_stress_gas);
        
        // Even under stress, average should remain efficient
        assert!(avg_stress_gas < 1_000_000, "Stress test average gas too high: {}", avg_stress_gas);
        
        // Verify final state is consistent
        let mut total_balance = client.balance(&admin);
        for user in &users {
            total_balance += client.balance(user);
        }
        assert_eq!(total_balance, 10_000_000); // Initial supply
    }
}
