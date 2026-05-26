// ─────────────────────────────────────────────────────────────────────────────
// RentRelay — MongoDB Schema Definitions
// Database: rentrelay_db
// All collections use MongoDB 7 with replica set (3 nodes)
// ─────────────────────────────────────────────────────────────────────────────

// ══════════════════════════════════════════════════════
// COLLECTION: users
// ══════════════════════════════════════════════════════
db.createCollection("users", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["user_id", "name", "email", "phone", "role", "created_at"],
      properties: {
        _id:           { bsonType: "objectId" },
        user_id:       { bsonType: "string",  description: "UUID v4" },
        name:          { bsonType: "string" },
        email:         { bsonType: "string" },
        phone:         { bsonType: "string",  pattern: "^[0-9]{10}$" },
        password_hash: { bsonType: "string",  description: "bcrypt hash" },
        role:          { bsonType: "string",  enum: ["TENANT", "LANDLORD", "ADMIN"] },
        aadhaar_hash:  { bsonType: "string",  description: "SHA-256 of Aadhaar number" },
        kyc_verified:  { bsonType: "bool" },
        kyc_verified_at: { bsonType: "date" },
        is_active:     { bsonType: "bool" },
        created_at:    { bsonType: "date" },
        updated_at:    { bsonType: "date" }
      }
    }
  }
});

db.users.createIndex({ "user_id": 1 }, { unique: true });
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "phone": 1 }, { unique: true });
db.users.createIndex({ "role": 1 });

// Sample document
// {
//   "_id": ObjectId("..."),
//   "user_id": "usr_01J8K2M9X4N7P3Q6R0S5T",
//   "name": "Ravi Kumar",
//   "email": "ravi@example.com",
//   "phone": "9876543210",
//   "password_hash": "$2b$12$...",
//   "role": "TENANT",
//   "aadhaar_hash": "e3b0c44298fc1c149afb...",
//   "kyc_verified": true,
//   "kyc_verified_at": ISODate("2025-01-15T10:00:00Z"),
//   "is_active": true,
//   "created_at": ISODate("2025-01-15T09:00:00Z"),
//   "updated_at": ISODate("2025-01-15T10:00:00Z")
// }


// ══════════════════════════════════════════════════════
// COLLECTION: properties
// ══════════════════════════════════════════════════════
db.createCollection("properties", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["property_id", "landlord_id", "title", "address", "city", "zone",
                 "rent_monthly", "deposit_amt", "is_available", "created_at"],
      properties: {
        property_id:   { bsonType: "string" },
        landlord_id:   { bsonType: "string" },
        title:         { bsonType: "string" },
        address:       { bsonType: "string" },
        city:          { bsonType: "string" },
        zone:          { bsonType: "string",  description: "Partition key for storage workers" },
        location: {
          bsonType: "object",
          description: "GeoJSON Point",
          required: ["type", "coordinates"],
          properties: {
            type:        { bsonType: "string", enum: ["Point"] },
            coordinates: { bsonType: "array",  description: "[longitude, latitude]" }
          }
        },
        bedrooms:      { bsonType: "int" },
        bathrooms:     { bsonType: "int" },
        area_sqft:     { bsonType: "double" },
        rent_monthly:  { bsonType: "double" },
        deposit_amt:   { bsonType: "double" },
        furnishing:    { bsonType: "string", enum: ["UNFURNISHED","SEMI_FURNISHED","FULLY_FURNISHED"] },
        amenities:     { bsonType: "array", items: { bsonType: "string" } },
        is_available:  { bsonType: "bool" },
        available_from:{ bsonType: "date" },
        images:        { bsonType: "array", items: { bsonType: "string" } },
        created_at:    { bsonType: "date" },
        updated_at:    { bsonType: "date" }
      }
    }
  }
});

db.properties.createIndex({ "property_id": 1 }, { unique: true });
db.properties.createIndex({ "landlord_id": 1 });
db.properties.createIndex({ "city": 1, "zone": 1 });
db.properties.createIndex({ "rent_monthly": 1 });
db.properties.createIndex({ "is_available": 1 });
db.properties.createIndex({ "location": "2dsphere" });   // geospatial queries


// ══════════════════════════════════════════════════════
// COLLECTION: lease_terms
// ══════════════════════════════════════════════════════
db.createCollection("lease_terms");

db.lease_terms.createIndex({ "property_id": 1 }, { unique: true });
db.lease_terms.createIndex({ "landlord_id": 1 });

// Sample document
// {
//   "property_id": "prop_01J8...",
//   "landlord_id": "usr_01J8...",
//   "lease_duration_mo": 11,
//   "notice_period_days": 30,
//   "preferred_tenant": "working_professional",
//   "allowed_types": ["family", "working_professional"],
//   "maintenance_charge": 500,
//   "payment_due_day": "5",
//   "updated_at": ISODate("2025-03-01T00:00:00Z")
// }


// ══════════════════════════════════════════════════════
// COLLECTION: rental_requests
// ══════════════════════════════════════════════════════
db.createCollection("rental_requests", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["request_id", "tenant_id", "preferred_city", "max_rent", "status", "created_at"],
      properties: {
        request_id:     { bsonType: "string" },
        tenant_id:      { bsonType: "string" },
        preferred_zone: { bsonType: "string" },
        preferred_city: { bsonType: "string" },
        bedrooms_needed:{ bsonType: "int" },
        max_rent:       { bsonType: "double" },
        furnishing:     { bsonType: "string" },
        move_in_date:   { bsonType: "date" },
        status:         {
          bsonType: "string",
          enum: ["OPEN", "MATCHED", "AGREEMENT_IN_PROGRESS", "CLOSED", "EXPIRED"]
        },
        matched_property_id: { bsonType: "string" },
        created_at:     { bsonType: "date" },
        expires_at:     { bsonType: "date" }
      }
    }
  }
});

db.rental_requests.createIndex({ "request_id": 1 }, { unique: true });
db.rental_requests.createIndex({ "tenant_id": 1 });
db.rental_requests.createIndex({ "preferred_city": 1, "preferred_zone": 1 });
db.rental_requests.createIndex({ "status": 1 });
db.rental_requests.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0 });  // TTL index


// ══════════════════════════════════════════════════════
// COLLECTION: agreements
// Central collection — also replicated via storage workers (Proj 2)
// ══════════════════════════════════════════════════════
db.createCollection("agreements", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["agreement_id", "tenant_id", "landlord_id", "property_id",
                 "state", "monthly_rent", "deposit_amount", "created_at"],
      properties: {
        agreement_id:    { bsonType: "string" },
        tenant_id:       { bsonType: "string" },
        landlord_id:     { bsonType: "string" },
        property_id:     { bsonType: "string" },
        state: {
          bsonType: "string",
          enum: ["CREATED","SIGNED","ESCROW_HELD","ACTIVE","DISPUTED",
                 "NOTICE_PERIOD","TERMINATING","COMPLETED","CANCELLED","RESOLVED"]
        },
        monthly_rent:    { bsonType: "double" },
        deposit_amount:  { bsonType: "double" },
        deposit_held:    { bsonType: "double",  description: "Current escrow balance" },
        lease_months:    { bsonType: "int" },
        notice_days:     { bsonType: "int" },
        document_hash:   { bsonType: "string",  description: "SHA-256 of signed PDF" },
        signatures: {
          bsonType: "array",
          items: {
            bsonType: "object",
            properties: {
              signer_id:      { bsonType: "string" },
              signature_hash: { bsonType: "string" },
              signed_at:      { bsonType: "date" }
            }
          }
        },
        start_date:      { bsonType: "date" },
        end_date:        { bsonType: "date" },
        // Proj 2 replication metadata
        worker_node:     { bsonType: "string" },
        replica_version: { bsonType: "int" },
        replica_nodes:   { bsonType: "array", items: { bsonType: "string" } },
        created_at:      { bsonType: "date" },
        updated_at:      { bsonType: "date" }
      }
    }
  }
});

db.agreements.createIndex({ "agreement_id": 1 }, { unique: true });
db.agreements.createIndex({ "tenant_id": 1 });
db.agreements.createIndex({ "landlord_id": 1 });
db.agreements.createIndex({ "property_id": 1 });
db.agreements.createIndex({ "state": 1 });
db.agreements.createIndex({ "tenant_id": 1, "state": 1 });
db.agreements.createIndex({ "worker_node": 1 });  // for Proj 2 shard queries


// ══════════════════════════════════════════════════════
// COLLECTION: agreement_events
// Append-only audit trail of every state transition
// ══════════════════════════════════════════════════════
db.createCollection("agreement_events");

db.agreement_events.createIndex({ "agreement_id": 1, "occurred_at": -1 });
db.agreement_events.createIndex({ "actor_id": 1 });
db.agreement_events.createIndex({ "occurred_at": -1 });

// Sample document
// {
//   "event_id":     "evt_01J9...",
//   "agreement_id": "agr_01J9...",
//   "old_state":    "ACTIVE",
//   "new_state":    "DISPUTED",
//   "actor_id":     "usr_landlord_01",
//   "notes":        "Tenant has not paid rent for 2 months",
//   "metadata":     { "evidence_docs": ["doc_01", "doc_02"] },
//   "occurred_at":  ISODate("2025-05-10T14:30:00Z")
// }


// ══════════════════════════════════════════════════════
// COLLECTION: payments
// ══════════════════════════════════════════════════════
db.createCollection("payments", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["receipt_id","agreement_id","payer_id","amount","payment_type","status","paid_at"],
      properties: {
        receipt_id:   { bsonType: "string" },
        agreement_id: { bsonType: "string" },
        payer_id:     { bsonType: "string" },
        amount:       { bsonType: "double" },
        payment_type: { bsonType: "string", enum: ["RENT","DEPOSIT","MAINTENANCE","PENALTY"] },
        status:       { bsonType: "string", enum: ["SUCCESS","FAILED","PENDING","REFUNDED"] },
        month_year:   { bsonType: "string",  description: "e.g. 2025-05" },
        upi_ref:      { bsonType: "string" },
        // Proj 2 quorum metadata
        replica_version: { bsonType: "int" },
        quorum_ack:      { bsonType: "int",  description: "Number of replicas that acked" },
        paid_at:      { bsonType: "date" }
      }
    }
  }
});

db.payments.createIndex({ "receipt_id": 1 }, { unique: true });
db.payments.createIndex({ "agreement_id": 1, "paid_at": -1 });
db.payments.createIndex({ "payer_id": 1 });
db.payments.createIndex({ "month_year": 1 });
db.payments.createIndex({ "status": 1 });


// ══════════════════════════════════════════════════════
// COLLECTION: disputes
// ══════════════════════════════════════════════════════
db.createCollection("disputes", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["dispute_id","agreement_id","raised_by","status","reason","raised_at"],
      properties: {
        dispute_id:        { bsonType: "string" },
        agreement_id:      { bsonType: "string" },
        raised_by:         { bsonType: "string" },
        status:            { bsonType: "string", enum: ["OPEN","UNDER_REVIEW","RESOLVED","DISMISSED"] },
        reason:            { bsonType: "string" },
        description:       { bsonType: "string" },
        resolution:        { bsonType: "string" },
        escrow_deducted:   { bsonType: "double" },
        evidence_doc_ids:  { bsonType: "array", items: { bsonType: "string" } },
        assigned_admin:    { bsonType: "string" },
        raised_at:         { bsonType: "date" },
        resolved_at:       { bsonType: "date" }
      }
    }
  }
});

db.disputes.createIndex({ "dispute_id": 1 }, { unique: true });
db.disputes.createIndex({ "agreement_id": 1 });
db.disputes.createIndex({ "raised_by": 1 });
db.disputes.createIndex({ "status": 1 });


// ══════════════════════════════════════════════════════
// COLLECTION: documents
// ══════════════════════════════════════════════════════
db.createCollection("documents", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["document_id","agreement_id","doc_type","storage_key","sha256_hash","uploaded_by","uploaded_at"],
      properties: {
        document_id:   { bsonType: "string" },
        agreement_id:  { bsonType: "string" },
        doc_type:      { bsonType: "string", enum: ["AGREEMENT","ID_PROOF","DISPUTE_EVIDENCE","RECEIPT","PHOTO"] },
        storage_key:   { bsonType: "string",  description: "MinIO/S3 object path" },
        sha256_hash:   { bsonType: "string" },
        size_bytes:    { bsonType: "long" },
        filename:      { bsonType: "string" },
        uploaded_by:   { bsonType: "string" },
        locked:        { bsonType: "bool",    description: "True when agreement is DISPUTED" },
        uploaded_at:   { bsonType: "date" }
      }
    }
  }
});

db.documents.createIndex({ "document_id": 1 }, { unique: true });
db.documents.createIndex({ "agreement_id": 1 });
db.documents.createIndex({ "sha256_hash": 1 });   // tamper check


// ══════════════════════════════════════════════════════
// COLLECTION: notifications
// ══════════════════════════════════════════════════════
db.createCollection("notifications");

db.notifications.createIndex({ "notification_id": 1 }, { unique: true });
db.notifications.createIndex({ "user_id": 1, "created_at": -1 });
db.notifications.createIndex({ "agreement_id": 1 });
db.notifications.createIndex({ "delivered": 1 });
// Auto-delete notifications older than 90 days
db.notifications.createIndex({ "created_at": 1 }, { expireAfterSeconds: 7776000 });


// ══════════════════════════════════════════════════════
// COLLECTION: storage_partition_table
// Maintained by the storage controller (Proj 2)
// ══════════════════════════════════════════════════════
db.createCollection("storage_partition_table");

db.storage_partition_table.createIndex({ "worker_id": 1 }, { unique: true });

// Sample document per worker
// {
//   "worker_id":      "storage-worker-1",
//   "worker_address": "storage-worker-1.rentrelay.svc.cluster.local:50060",
//   "shard_start":    0,
//   "shard_end":      63,
//   "is_alive":       true,
//   "last_heartbeat": ISODate("2025-05-22T10:00:00Z"),
//   "stored_keys":    12450,
//   "disk_usage_pct": 34.2,
//   "replica_for":    ["storage-worker-2", "storage-worker-3"]
// }


// ══════════════════════════════════════════════════════
// COLLECTION: storage_replication_log
// Tracks async replication events for debugging (Proj 2)
// ══════════════════════════════════════════════════════
db.createCollection("storage_replication_log");

db.storage_replication_log.createIndex({ "key": 1, "created_at": -1 });
db.storage_replication_log.createIndex({ "from_worker": 1 });
db.storage_replication_log.createIndex({ "to_worker": 1 });
// Auto-delete replication logs older than 7 days
db.storage_replication_log.createIndex({ "created_at": 1 }, { expireAfterSeconds: 604800 });

// Sample document
// {
//   "key":             "agr_01J9...",
//   "from_worker":     "storage-worker-1",
//   "to_worker":       "storage-worker-3",
//   "replica_version": 7,
//   "event":           "ASYNC_REPLICATE",   // or "REBALANCE" | "FAILURE_RECOVERY"
//   "success":         true,
//   "duration_ms":     12,
//   "created_at":      ISODate("2025-05-22T10:00:05Z")
// }
