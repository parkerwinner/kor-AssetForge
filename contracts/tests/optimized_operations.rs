use soroban_sdk::{contract, contractimpl, Address, Env, testutils::Address as _, testutils::Ledger};
use kor_assetforge_contracts::AssetTokenClient;

#[cfg(test)]
mod tests {
    use super::*;
    use soroban_sdk::testutils::{Address as _, Ledger};

    fn setup(env: &Env) -> (AssetTokenClient<'_>, Address, Address) {
        let contract_id = env.register_contract(None, kor_assetforge_contracts::AssetToken);
        let client = AssetTokenClient::new(env, &contract_id);
        
        let ec_addr = env.register_contract(None, kor_assetforge_contracts::EmergencyControl);
        let ec_client = kor_assetforge_contracts::EmergencyControlClient::new(env, &ec_addr);
        
        let admin = Address::generate(env);
        ec_client.initialize(&admin);
        
        let name = soroban_sdk::String::from_str(env, "Optimized Asset Token");
        let symbol = soroban_sdk::String::from_str(env, "OAT");
        let supply = 1_000_000;
        
        client.initialize(&admin, &name, &symbol, &7, &supply);
        (client, admin, ec_addr)
    }

    #[test]
    fn test_optimized_transfer_basic() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user1 = Address::generate(&env);
        let user2 = Address::generate(&env);
        
        // Setup initial balance
        client.transfer(&admin, &user1, &1000, &1, &ec_id);
        
        let balance_before_user1 = client.balance(&user1);
        let balance_before_user2 = client.balance(&user2);
        
        // Perform transfer
        client.transfer(&user1, &user2, &500, &1, &ec_id);
        
        // Verify balances
        assert_eq!(client.balance(&user1), balance_before_user1 - 500);
        assert_eq!(client.balance(&user2), balance_before_user2 + 500);
    }

    #[test]
    fn test_optimized_transfer_insufficient_balance() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user1 = Address::generate(&env);
        let user2 = Address::generate(&env);
        
        // Setup initial balance
        client.transfer(&admin, &user1, &100, &1, &ec_id);
        
        // Attempt transfer with insufficient balance should panic
        std::panic::catch_unwind(|| {
            client.transfer(&user1, &user2, &200, &1, &ec_id);
        }).unwrap_err();
    }

    #[test]
    fn test_optimized_transfer_from_basic() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user1 = Address::generate(&env);
        let user2 = Address::generate(&env);
        
        // Setup initial balance
        client.transfer(&admin, &user1, &1000, &1, &ec_id);
        
        let balance_before_user1 = client.balance(&user1);
        let balance_before_user2 = client.balance(&user2);
        
        // Perform transfer_from
        client.transfer_from(&user1, &user1, &user2, &500, &1, &ec_id);
        
        // Verify balances
        assert_eq!(client.balance(&user1), balance_before_user1 - 500);
        assert_eq!(client.balance(&user2), balance_before_user2 + 500);
    }

    #[test]
    fn test_optimized_mint_basic() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user = Address::generate(&env);
        
        let balance_before = client.balance(&user);
        let supply_before = client.total_supply();
        
        // Perform mint
        client.mint(&user, &1000, &1, &ec_id);
        
        // Verify balance and supply
        assert_eq!(client.balance(&user), balance_before + 1000);
        assert_eq!(client.total_supply(), supply_before + 1000);
    }

    #[test]
    fn test_optimized_burn_basic() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user = Address::generate(&env);
        
        // Setup initial balance
        client.transfer(&admin, &user, &1000, &1, &ec_id);
        
        let balance_before = client.balance(&user);
        let supply_before = client.total_supply();
        
        // Perform burn
        client.burn(&user, &500, &1, &ec_id);
        
        // Verify balance and supply
        assert_eq!(client.balance(&user), balance_before - 500);
        assert_eq!(client.total_supply(), supply_before - 500);
    }

    #[test]
    fn test_optimized_burn_insufficient_balance() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user = Address::generate(&env);
        
        // Setup initial balance
        client.transfer(&admin, &user, &100, &1, &ec_id);
        
        // Attempt burn with insufficient balance should panic
        std::panic::catch_unwind(|| {
            client.burn(&user, &200, &1, &ec_id);
        }).unwrap_err();
    }

    #[test]
    fn test_optimized_operations_preserve_total_supply() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user1 = Address::generate(&env);
        let user2 = Address::generate(&env);
        
        let initial_supply = client.total_supply();
        
        // Setup initial balances
        client.transfer(&admin, &user1, &1000, &1, &ec_id);
        client.transfer(&admin, &user2, &1000, &1, &ec_id);
        
        // Transfer should not affect total supply
        client.transfer(&user1, &user2, &500, &1, &ec_id);
        assert_eq!(client.total_supply(), initial_supply);
        
        // Mint should increase total supply
        client.mint(&user1, &1000, &1, &ec_id);
        assert_eq!(client.total_supply(), initial_supply + 1000);
        
        // Burn should decrease total supply
        client.burn(&user2, &500, &1, &ec_id);
        assert_eq!(client.total_supply(), initial_supply + 500);
    }

    #[test]
    fn test_optimized_arithmetic_safety() {
        let env = Env::default();
        env.mock_all_auths();
        let (client, admin, ec_id) = setup(&env);
        
        let user = Address::generate(&env);
        
        // Test large amounts to ensure arithmetic safety
        client.mint(&user, &9223372036854775807, &1, &ec_id); // Max i128
        
        // Test transfer with large amount
        client.transfer(&user, &admin, &4611686018427387903, &1, &ec_id);
        
        // Verify balances are correct
        assert_eq!(client.balance(&user), 4611686018427387904);
        assert_eq!(client.balance(&admin), 1000000 + 4611686018427387903);
    }

    #[test]
    fn test_optimized_authorization_enforcement() {
        let env = Env::default();
        // Don't mock all auths to test authorization
        let (client, admin, ec_id) = setup(&env);
        
        let user1 = Address::generate(&env);
        let user2 = Address::generate(&env);
        
        // Setup initial balance
        client.transfer(&admin, &user1, &1000, &1, &ec_id);
        
        // Attempt transfer without proper auth should fail
        std::panic::catch_unwind(|| {
            client.transfer(&user2, &user1, &500, &1, &ec_id);
        }).unwrap_err();
        
        // Attempt burn without proper auth should fail
        std::panic::catch_unwind(|| {
            client.burn(&user2, &500, &1, &ec_id);
        }).unwrap_err();
    }
}
