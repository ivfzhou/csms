# Copyright (c) 2025 ivfzhou
# csms is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

# fastlane version is 2.227.2

lane :get_bundle_info do |options|
  bundle_id = options[:bundle_id]
  team_id = options[:team_id]
  require 'spaceship'
  Spaceship::Portal.login
  Spaceship::Portal.select_team(team_id: team_id)
  app = Spaceship::Portal.app.find(bundle_id)
  if app
    require 'json'
    puts JSON.generate(
      {
        bundleId: app.bundle_id,
        name: app.name,
        id: app.app_id,
        platform: app.platform,
        capabilities: app.enable_services,
        teamId: app.prefix,
        isWildcard: app.is_wildcard,
        features: app.features,
        devPushEnabled: app.dev_push_enabled,
        prodPushEnabled: app.prod_push_enabled,
        appGroupsCount: app.app_groups_count,
        cloudContainersCount: app.cloud_containers_count,
        identifiersCount: app.identifiers_count,
        associatedGroups: app.associated_groups,
        associatedCloudContainers: app.associated_cloud_containers
      }
    )
  else
    puts 'bundle id not found'
  end
end

lane :del_bundle do |options|
  bundle_id = options[:bundle_id]
  team_id = options[:team_id]
  require 'spaceship'
  Spaceship::Portal.login
  Spaceship::Portal.select_team(team_id: team_id)
  app = Spaceship::Portal.app.find(bundle_id)
  if app
    app.delete!
  else
    puts 'bundle id not found'
  end
end

lane :get_profile_info do |options|
  bundle_id = options[:bundle_id]
  team_id = options[:team_id]
  filename = options[:filename]
  require 'spaceship'
  Spaceship::Portal.login
  Spaceship::Portal.select_team(team_id: team_id)
  target = Spaceship::Portal.provisioning_profile.all.find do |profile|
    profile.app.bundle_id == bundle_id &&
      profile.app.prefix == team_id &&
      profile.name == filename
  end
  if target
    puts JSON.generate(
      {
        id: target.id,
        uuid: target.uuid,
        expires: target.expires,
        name: target.name,
        status: target.status,
        type: target.type,
        platform: target.platform,
        certificateId: target.certificates.first&.id,
        teamId: target.app.prefix
      }
    )
  else
    puts 'profile not found'
  end
end

lane :del_profile do |options|
  id = options[:id]
  team_id = options[:team_id]
  require 'spaceship'
  Spaceship::Portal.login
  Spaceship::Portal.select_team(team_id: team_id)
  target = Spaceship::Portal.provisioning_profile.all.find do |profile|
    profile.id == id
  end
  if target
    target.delete!
  else
    puts 'profile not found'
  end
end

lane :get_cert_info do |options|
  bundle_id = options[:bundle_id]
  team_id = options[:team_id]
  expire = options[:expire]
  name_like = options[:name_like]
  require 'spaceship'
  Spaceship::Portal.login
  Spaceship::Portal.select_team(team_id: team_id)
  target = Spaceship::Portal.certificate.all.select do |cert|
    cert.owner_type == 'bundle' && cert.owner_name == bundle_id && cert.expires.strftime('%Y%m%d%H%M%S') == expire &&
      cert.name.include?(name_like)
  end
  list = target.map do |v|
    {
      id: v.id,
      name: v.name,
      status: v.status,
      created: v.created,
      expires: v.expires,
      ownerId: v.owner_id,
      typeDisplayId: v.type_display_id,
      canDownload: v.can_download,
      ownerType: v.owner_type,
      ownerName: v.owner_name
    }
  end
  if list
    puts JSON.generate(list)
  else
    puts 'certificate not found'
  end
end

lane :del_cert do |options|
  bundle_id = options[:bundle_id]
  id = options[:id]
  team_id = options[:team_id]
  name_like = options[:name_like]
  require 'spaceship'
  Spaceship::Portal.login
  Spaceship::Portal.select_team(team_id: team_id)
  target = Spaceship::Portal.certificate.all.find do |cert|
    cert.id == id && cert.is_push? && cert.owner_type == 'bundle' && cert.owner_name == bundle_id &&
      cert.name.include?(name_like)
  end
  if target
    target.revoke!
  else
    puts 'certificate not found'
  end
end
