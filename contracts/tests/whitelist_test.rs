#[cfg(test)]
mod test {
    use soroban_sdk::testutils::{Address as _, Ledger as _};
    use soroban_sdk::{Address, Env, Vec};

    use kor_assetforge_contracts::whitelist::{
        AccreditedTier, Whitelist, WhitelistClient,
    };

    fn setup() -> (Env, WhitelistClient<'static>, Address) {
        let env = Env::default();
        env.mock_all_auths();
        let contract_id = env.register_contract(None, Whitelist);
        let client = WhitelistClient::new(&env, &contract_id);
        let admin = Address::generate(&env);
        client.initialize(&admin);
        (env, client, admin)
    }

    #[test]
    fn test_initialize() {
        let env = Env::default();
        env.mock_all_auths();
        let contract_id = env.register_contract(None, Whitelist);
        let client = WhitelistClient::new(&env, &contract_id);
        let admin = Address::generate(&env);
        client.initialize(&admin);
        assert_eq!(client.get_user_count(), 0);
    }

    #[test]
    #[should_panic(expected = "already initialized")]
    fn test_initialize_twice_panics() {
        let env = Env::default();
        env.mock_all_auths();
        let contract_id = env.register_contract(None, Whitelist);
        let client = WhitelistClient::new(&env, &contract_id);
        let admin = Address::generate(&env);
        client.initialize(&admin);
        client.initialize(&admin);
    }

    #[test]
    fn test_add_entry() {
        let (env, client, admin) = setup();
        let user = Address::generate(&env);
        client.add_entry(&admin, &user, &AccreditedTier::Tier1);
        assert!(client.is_whitelisted(&user));
        assert_eq!(client.get_user_count(), 1);
    }

    #[test]
    fn test_add_entry_all_tiers() {
        let (env, client, admin) = setup();
        let u1 = Address::generate(&env);
        let u2 = Address::generate(&env);
        let u3 = Address::generate(&env);
        client.add_entry(&admin, &u1, &AccreditedTier::Tier1);
        client.add_entry(&admin, &u2, &AccreditedTier::Tier2);
        client.add_entry(&admin, &u3, &AccreditedTier::Tier3);
        assert_eq!(client.get_tier(&u1).unwrap(), AccreditedTier::Tier1);
        assert_eq!(client.get_tier(&u2).unwrap(), AccreditedTier::Tier2);
        assert_eq!(client.get_tier(&u3).unwrap(), AccreditedTier::Tier3);
    }

    #[test]
    fn test_get_entry() {
        let (env, client, admin) = setup();
        let user = Address::generate(&env);
        client.add_entry(&admin, &user, &AccreditedTier::Tier2);
        let entry = client.get_entry(&user).unwrap();
        assert_eq!(entry.user, user);
        assert_eq!(entry.tier, AccreditedTier::Tier2);
        assert!(entry.active);
        assert!(entry.expires_at > entry.added_at);
    }

    #[test]
    fn test_get_entry_not_found() {
        let (env, client, _admin) = setup();
        let user = Address::generate(&env);
        let entry = client.get_entry(&user);
        assert!(entry.is_none());
    }

    #[test]
    fn test_remove_entry() {
        let (env, client, admin) = setup();
        let user = Address::generate(&env);
        client.add_entry(&admin, &user, &AccreditedTier::Tier1);
        assert!(client.is_whitelisted(&user));
        client.remove_entry(&admin, &user);
        assert!(!client.is_whitelisted(&user));
    }

    #[test]
    fn test_expiry_after_one_year() {
        let (env, client, admin) = setup();
        let user = Address::generate(&env);
        env.ledger().with_mut(|l| l.timestamp = 1_000_000);
        client.add_entry(&admin, &user, &AccreditedTier::Tier1);
        assert!(client.is_whitelisted(&user));
        env.ledger().with_mut(|l| l.timestamp = 1_000_000 + 31_536_001);
        assert!(!client.is_whitelisted(&user));
    }

    #[test]
    fn test_expiry_boundary() {
        let (env, client, admin) = setup();
        let user = Address::generate(&env);
        env.ledger().with_mut(|l| l.timestamp = 100);
        client.add_entry(&admin, &user, &AccreditedTier::Tier1);
        env.ledger().with_mut(|l| l.timestamp = 100 + 31_536_000);
        assert!(!client.is_whitelisted(&user));
    }

    #[test]
    fn test_bulk_add_entries() {
        let (env, client, admin) = setup();
        let u1 = Address::generate(&env);
        let u2 = Address::generate(&env);
        let u3 = Address::generate(&env);
        let users = Vec::from_array(&env, [u1.clone(), u2.clone(), u3.clone()]);
        let tiers = Vec::from_array(&env, [AccreditedTier::Tier1, AccreditedTier::Tier2, AccreditedTier::Tier3]);
        client.add_bulk_entries(&admin, &users, &tiers);
        assert!(client.is_whitelisted(&u1));
        assert!(client.is_whitelisted(&u2));
        assert!(client.is_whitelisted(&u3));
        assert_eq!(client.get_user_count(), 3);
    }

    #[test]
    fn test_remove_bulk_entries() {
        let (env, client, admin) = setup();
        let u1 = Address::generate(&env);
        let u2 = Address::generate(&env);
        let users = Vec::from_array(&env, [u1.clone(), u2.clone()]);
        let tiers = Vec::from_array(&env, [AccreditedTier::Tier1, AccreditedTier::Tier2]);
        client.add_bulk_entries(&admin, &users, &tiers);
        assert!(client.is_whitelisted(&u1));
        assert!(client.is_whitelisted(&u2));
        client.remove_bulk_entries(&admin, &users);
        assert!(!client.is_whitelisted(&u1));
        assert!(!client.is_whitelisted(&u2));
    }

    #[test]
    fn test_bulk_add_skips_duplicates() {
        let (env, client, admin) = setup();
        let user = Address::generate(&env);
        client.add_entry(&admin, &user, &AccreditedTier::Tier1);
        let users = Vec::from_array(&env, [user.clone()]);
        let tiers = Vec::from_array(&env, [AccreditedTier::Tier2]);
        client.add_bulk_entries(&admin, &users, &tiers);
        assert_eq!(client.get_user_count(), 1);
    }

    #[test]
    fn test_refresh_entry() {
        let (env, client, admin) = setup();
        let user = Address::generate(&env);
        env.ledger().with_mut(|l| l.timestamp = 100);
        client.add_entry(&admin, &user, &AccreditedTier::Tier1);
        let entry_before = client.get_entry(&user).unwrap();
        let original_expiry = entry_before.expires_at;
        env.ledger().with_mut(|l| l.timestamp = 100 + 15_768_000);
        assert!(client.is_whitelisted(&user));
        client.refresh_entry(&admin, &user);
        let entry_after = client.get_entry(&user).unwrap();
        assert!(entry_after.expires_at > original_expiry);
    }

    #[test]
    fn test_get_tier_not_whitelisted() {
        let (env, client, _admin) = setup();
        let user = Address::generate(&env);
        let tier = client.get_tier(&user);
        assert!(tier.is_none());
    }

    #[test]
    fn test_get_tier_expired() {
        let (env, client, admin) = setup();
        let user = Address::generate(&env);
        env.ledger().with_mut(|l| l.timestamp = 100);
        client.add_entry(&admin, &user, &AccreditedTier::Tier2);
        env.ledger().with_mut(|l| l.timestamp = 100 + 31_536_001);
        let tier = client.get_tier(&user);
        assert!(tier.is_none());
    }

    #[test]
    fn test_get_all_users() {
        let (env, client, admin) = setup();
        let u1 = Address::generate(&env);
        let u2 = Address::generate(&env);
        client.add_entry(&admin, &u1, &AccreditedTier::Tier1);
        client.add_entry(&admin, &u2, &AccreditedTier::Tier2);
        let users = client.get_all_users();
        assert_eq!(users.len(), 2);
    }

    #[test]
    fn test_user_count() {
        let (env, client, admin) = setup();
        assert_eq!(client.get_user_count(), 0);
        let user = Address::generate(&env);
        client.add_entry(&admin, &user, &AccreditedTier::Tier1);
        assert_eq!(client.get_user_count(), 1);
    }

    #[test]
    fn test_add_same_user_after_expiry() {
        let (env, client, admin) = setup();
        let user = Address::generate(&env);
        env.ledger().with_mut(|l| l.timestamp = 100);
        client.add_entry(&admin, &user, &AccreditedTier::Tier1);
        env.ledger().with_mut(|l| l.timestamp = 100 + 31_536_001);
        assert!(!client.is_whitelisted(&user));
        client.add_entry(&admin, &user, &AccreditedTier::Tier2);
        assert!(client.is_whitelisted(&user));
        assert_eq!(client.get_user_count(), 1);
    }

    #[test]
    fn test_is_whitelisted_no_contract() {
        let (_env, client, _admin) = setup();
        let user = Address::generate(&_env);
        let result = client.is_whitelisted(&user);
        assert!(!result);
    }

    #[test]
    #[should_panic(expected = "not authorized")]
    fn test_non_admin_cannot_add() {
        let (env, client, _admin) = setup();
        let non_admin = Address::generate(&env);
        let user = Address::generate(&env);
        env.mock_all_auths();
        client.add_entry(&non_admin, &user, &AccreditedTier::Tier1);
    }

    #[test]
    #[should_panic(expected = "not authorized")]
    fn test_non_admin_cannot_remove() {
        let (env, client, _admin) = setup();
        let non_admin = Address::generate(&env);
        let user = Address::generate(&env);
        env.mock_all_auths();
        client.remove_entry(&non_admin, &user);
    }

    #[test]
    fn test_marketplace_integration_deployed_together() {
        use kor_assetforge_contracts::marketplace::{Marketplace, MarketplaceClient};
        use kor_assetforge_contracts::emergency_control::{EmergencyControl, EmergencyControlClient};

        let env = Env::default();
        env.mock_all_auths();

        let admin = Address::generate(&env);

        let ec_id = env.register_contract(None, EmergencyControl);
        let ec_client = EmergencyControlClient::new(&env, &ec_id);
        ec_client.initialize(&admin);

        let mp_id = env.register_contract(None, Marketplace);
        let mp_client = MarketplaceClient::new(&env, &mp_id);
        mp_client.initialize(&admin);

        let wl_id = env.register_contract(None, Whitelist);
        let wl_client = WhitelistClient::new(&env, &wl_id);
        wl_client.initialize(&admin);

        mp_client.set_whitelist_contract(&admin, &wl_id);

        let seller = Address::generate(&env);
        wl_client.add_entry(&admin, &seller, &AccreditedTier::Tier1);
        assert!(wl_client.is_whitelisted(&seller));

        let lid = mp_client.create_listing(&seller, &1, &100, &1000, &ec_id, &None);
        assert_eq!(lid, 1);

        let buyer = Address::generate(&env);
        let result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
            mp_client.purchase(&buyer, &1, &50, &1, &ec_id);
        }));
        assert!(result.is_err());
    }

    #[test]
    fn test_marketplace_integration_whitelisted_buyer() {
        use kor_assetforge_contracts::marketplace::{Marketplace, MarketplaceClient};
        use kor_assetforge_contracts::emergency_control::{EmergencyControl, EmergencyControlClient};

        let env = Env::default();
        env.mock_all_auths();

        let admin = Address::generate(&env);

        let ec_id = env.register_contract(None, EmergencyControl);
        let ec_client = EmergencyControlClient::new(&env, &ec_id);
        ec_client.initialize(&admin);

        let mp_id = env.register_contract(None, Marketplace);
        let mp_client = MarketplaceClient::new(&env, &mp_id);
        mp_client.initialize(&admin);

        let wl_id = env.register_contract(None, Whitelist);
        let wl_client = WhitelistClient::new(&env, &wl_id);
        wl_client.initialize(&admin);

        mp_client.set_whitelist_contract(&admin, &wl_id);

        let seller = Address::generate(&env);
        let buyer = Address::generate(&env);

        wl_client.add_entry(&admin, &seller, &AccreditedTier::Tier2);
        wl_client.add_entry(&admin, &buyer, &AccreditedTier::Tier3);

        mp_client.create_listing(&seller, &10, &100, &500, &ec_id, &None);
        let result = mp_client.purchase(&buyer, &1, &50, &10, &ec_id);
        assert!(result);
    }

    #[test]
    fn test_marketplace_no_whitelist_contract_configured() {
        use kor_assetforge_contracts::marketplace::{Marketplace, MarketplaceClient};
        use kor_assetforge_contracts::emergency_control::{EmergencyControl, EmergencyControlClient};

        let env = Env::default();
        env.mock_all_auths();

        let admin = Address::generate(&env);

        let ec_id = env.register_contract(None, EmergencyControl);
        let ec_client = EmergencyControlClient::new(&env, &ec_id);
        ec_client.initialize(&admin);

        let mp_id = env.register_contract(None, Marketplace);
        let mp_client = MarketplaceClient::new(&env, &mp_id);
        mp_client.initialize(&admin);

        let seller = Address::generate(&env);
        let buyer = Address::generate(&env);

        mp_client.create_listing(&seller, &1, &100, &1000, &ec_id, &None);
        let result = mp_client.purchase(&buyer, &1, &50, &1, &ec_id);
        assert!(result);
    }
}
