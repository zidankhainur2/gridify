import { Users } from "lucide-react";
import { GroupData } from "../types";

export default function BRGroupList({ groups }: { groups: GroupData[] }) {
  if (!groups || groups.length === 0) return null;

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <div className="p-1.5 bg-surface-elevated rounded-lg border border-border-default">
          <Users size={18} className="text-secondary" />
        </div>
        <h3 className="text-base font-bold text-text-primary">Distribusi Grup Lobby</h3>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {groups.map((group) => (
          <div
            key={group.name}
            className="flex flex-col bg-surface-raised border border-border-default rounded-xl overflow-hidden shadow-xs"
          >
            {/* Group Header */}
            <div className="px-4 py-3 bg-surface-elevated border-b border-border-subtle flex justify-between items-center">
              <span className="font-bold text-sm text-text-primary">Grup {group.name}</span>
              <span className="text-xs bg-surface-base text-text-muted px-2 py-0.5 rounded-full border border-border-default font-mono">
                {group.teams.length} Tim
              </span>
            </div>

            {/* Teams List */}
            <div className="p-2 flex-1">
              <ul className="space-y-1">
                {group.teams.map((team, idx) => (
                  <li
                    key={team.id}
                    className="flex items-center gap-3 p-2 rounded-lg hover:bg-surface-elevated/50 transition-colors text-xs text-text-secondary"
                  >
                    <span className="w-5 h-5 flex items-center justify-center bg-surface-base border border-border-subtle rounded text-[10px] text-text-muted font-mono">
                      {idx + 1}
                    </span>
                    <span className="truncate font-medium text-text-primary">{team.name}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
