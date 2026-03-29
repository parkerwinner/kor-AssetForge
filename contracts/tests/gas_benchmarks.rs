use soroban_sdk::{contract, contractimpl, Address, Env, testutils::Address as _, testutils::Ledger};
use kor_assetforge_contracts::AssetTokenClient;

#[contract]
pub struct GasBenchmark;

#[contractimpl]
impl GasBenchmark {
    /// Benchmark transfer operation gas usage
    pub fn benchmark_transfer(env: Env, asset_id: u64, emergency_control_id: Address) -> u64 {
        let start_gas = env.budget().gas_consumed();
        
        let admin = Address::generate(&env);
        let user1 = Address::generate(&env);
        let user2 = Address::generate(&env);
        
        // Initialize contract
        let contract_id = env.register_contract(None, kor_assetforge_contracts::AssetToken);
        let client = AssetTokenClient::new(&env, &contract_id);
        
        // Setup initial balances
        client.transfer(&admin, &user1, &1000, &asset_id, &emergency_control_id);
        
        // Measure transfer gas
        client.transfer(&user1, &user2, &500, &asset_id, &emergency_control_id);
        
        let end_gas = env.budget().gas_consumed();
        end_gas - start_gas
    }
    
    /// Benchmark transfer_from operation gas usage
    pub fn benchmark_transfer_from(env: Env, asset_id: u64, emergency_control_id: Address) -> u64 {
        let start_gas = env.budget().gas_consumed();
        
        let admin = Address::generate(&env);
        let user1 = Address::generate(&env);
        let user2 = Address::generate(&env);
        
        let contract_id = env.register_contract(None, kor_assetforge_contracts::AssetToken);
        let client = AssetTokenClient::new(&env, &contract_id);
        
        // Setup initial balances
        client.transfer(&admin, &user1, &1000, &asset_id, &emergency_control_id);
        
        // Measure transfer_from gas
        client.transfer_from(&user1, &user1, &user2, &500, &asset_id, &emergency_control_id);
        
        let end_gas = env.budget().gas_consumed();
        end_gas - start_gas
    }
    
    /// Benchmark mint operation gas usage
    pub fn benchmark_mint(env: Env, asset_id: u64, emergency_control_id: Address) -> u64 {
        let start_gas = env.budget().gas_consumed();
        
        let admin = Address::generate(&env);
        let user = Address::generate(&env);
        
        let contract_id = env.register_contract(None, kor_assetforge_contracts::AssetToken);
        let client = AssetTokenClient::new(&env, &contract_id);
        
        // Measure mint gas
        client.mint(&user, &1000, &asset_id, &emergency_control_id);
        
        let end_gas = env.budget().gas_consumed();
        end_gas - start_gas
    }
    
    /// Benchmark burn operation gas usage
    pub fn benchmark_burn(env: Env, asset_id: u64, emergency_control_id: Address) -> u64 {
        let start_gas = env.budget().gas_consumed();
        
        let admin = Address::generate(&env);
        let user = Address::generate(&env);
        
        let contract_id = env.register_contract(None, kor_assetforge_contracts::AssetToken);
        let client = AssetTokenClient::new(&env, &contract_id);
        
        // Setup initial balance
        client.transfer(&admin, &user, &1000, &asset_id, &emergency_control_id);
        
        // Measure burn gas
        client.burn(&user, &500, &asset_id, &emergency_control_id);
        
        let end_gas = env.budget().gas_consumed();
        end_gas - start_gas
    }
    
    /// Run comprehensive benchmark suite
    pub fn run_all_benchmarks(env: Env, asset_id: u64, emergency_control_id: Address) -> (u64, u64, u64, u64) {
        let transfer_gas = Self::benchmark_transfer(env.clone(), asset_id, emergency_control_id.clone());
        let transfer_from_gas = Self::benchmark_transfer_from(env.clone(), asset_id, emergency_control_id.clone());
        let mint_gas = Self::benchmark_mint(env.clone(), asset_id, emergency_control_id.clone());
        let burn_gas = Self::benchmark_burn(env.clone(), asset_id, emergency_control_id);
        
        (transfer_gas, transfer_from_gas, mint_gas, burn_gas)
    }
}
